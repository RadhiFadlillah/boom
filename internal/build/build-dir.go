package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"
	fp "path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-boom/boom/internal/model"
)

// buildDir builds directory for specified URL path.
func (wk *Worker) buildDir(urlPath string, w io.Writer) ([]string, error) {
	// Fetch page number from URL
	pageNumber := -1
	cleanURLPath := urlPath

	urlPathBase := path.Base(urlPath)
	if isNum, number := isNumber(urlPathBase); isNum {
		pageNumber = number
		cleanURLPath = path.Dir(cleanURLPath)
	}

	// Now since the URL path clean from page number, we can generate
	// file path to process.
	dirPath := fp.Join(wk.ContentDir, cleanURLPath)
	indexMdPath := fp.Join(wk.ContentDir, cleanURLPath, "_index.md")
	if !isDir(dirPath) {
		return nil, fmt.Errorf("%s is not part of site content", urlPath)
	}

	// Parse metadata
	meta, content, err := wk.parsePath(indexMdPath)
	if err != nil {
		return nil, err
	}

	// Create template data
	tplData := model.DirTemplate{
		URLPath:     path.Join("/", urlPath),
		Title:       meta.Title,
		Description: meta.Description,
		Author:      meta.Author,
		PageSize:    meta.Pagination,
	}

	// Set content
	if !meta.Draft || (meta.Draft && wk.buildDraft) {
		tplData.Content = content
	}

	// Create path trails
	trailSegments := strings.Split(cleanURLPath, "/")
	if cleanURLPath != "" && cleanURLPath != "." && len(trailSegments) > 0 {
		trailSegments = append([]string{"."}, trailSegments...)
	}

	for i := 1; i < len(trailSegments); i++ {
		parentPath := strings.Join(trailSegments[:i], "/")
		parentFilePath := fp.Join(wk.ContentDir, parentPath, "_index.md")
		parentMeta, _, err := wk.parsePath(parentFilePath)
		if err != nil {
			return nil, err
		}

		tplData.PathTrails = append(tplData.PathTrails, model.ContentPath{
			URLPath: path.Join("/", parentPath),
			Title:   parentMeta.Title,
			IsDir:   true,
		})
	}

	tplData.PathTrails = append(tplData.PathTrails, model.ContentPath{
		URLPath: path.Join("/", cleanURLPath),
		Title:   tplData.Title,
		IsDir:   true,
	})

	// Fetch child items
	items, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	subDirs := []model.ContentPath{}
	subFiles := []model.ContentPath{}
	for _, item := range items {
		itemName := item.Name()
		itemExt := fp.Ext(itemName)
		itemPath := fp.Join(dirPath, itemName)
		itemURLPath := path.Join(cleanURLPath, strings.TrimSuffix(itemName, itemExt))

		itemMeta, _, err := wk.parsePath(itemPath)
		if err != nil {
			return nil, err
		}

		if item.IsDir() {
			subDirItems, err := ioutil.ReadDir(itemPath)
			if err != nil {
				return nil, err
			}

			nChild := 0
			for _, subItem := range subDirItems {
				subItemName := subItem.Name()
				if subItem.IsDir() {
					nChild++
					continue
				}

				if fp.Ext(subItemName) != ".md" || subItemName == "_index.md" {
					continue
				}

				subItemPath := fp.Join(itemPath, subItemName)
				subItemMeta, _, _ := wk.parsePath(subItemPath)
				if !subItemMeta.Draft || (subItemMeta.Draft && wk.buildDraft) {
					nChild++
				}
			}

			subDirs = append(subDirs, model.ContentPath{
				IsDir:   true,
				URLPath: path.Join("/", itemURLPath),
				Title:   itemMeta.Title,
				NChild:  nChild,
			})
			continue
		}

		if !item.IsDir() && itemExt == ".md" && itemName != "_index.md" {
			if itemMeta.Draft && !wk.buildDraft {
				continue
			}

			itemTime := itemMeta.UpdateTime
			if itemTime.IsZero() {
				itemTime = itemMeta.CreateTime
			}

			subFiles = append(subFiles, model.ContentPath{
				URLPath:    path.Join("/", itemURLPath),
				Title:      itemMeta.Title,
				UpdateTime: itemTime,
			})
		}
	}

	// Sort items
	sort.Slice(subDirs, func(a, b int) bool {
		titleA := subDirs[a].Title
		titleB := subDirs[b].Title
		return strings.ToLower(titleA) < strings.ToLower(titleB)
	})

	sort.Slice(subFiles, func(a, b int) bool {
		timeA := subFiles[a].UpdateTime
		timeB := subFiles[b].UpdateTime
		return timeA.After(timeB)
	})

	// Merge sub dirs and sub files
	dirItems := append(subDirs, subFiles...)

	// Fetch all tags within active directory
	mapDirTags := make(map[string]int)
	fnWalk := func(path string, info os.FileInfo, err error) error {
		// We look for markdown file
		if info.IsDir() || fp.Ext(path) != ".md" || fp.Base(path) == "_index.md" {
			return nil
		}

		// Parse file
		fileMeta, _, err := wk.parsePath(path)
		if err != nil {
			return err
		}

		// Save tags
		for _, tag := range fileMeta.Tags {
			mapDirTags[tag]++
		}
		return nil
	}

	err = fp.Walk(dirPath, fnWalk)
	if err != nil {
		return nil, err
	}

	// Sort tags
	dirTags := []model.TagPath{}
	for tag, count := range mapDirTags {
		dirTags = append(dirTags, model.TagPath{
			URLPath: path.Join("/", cleanURLPath, "tag-"+tag),
			Name:    tag,
			Count:   count,
		})
	}

	sort.Slice(dirTags, func(a, b int) bool {
		nameA := dirTags[a].Name
		nameB := dirTags[b].Name
		return strings.ToLower(nameA) < strings.ToLower(nameB)
	})

	tplData.ChildTags = dirTags

	// Calculate pagination stuffs
	if tplData.PageSize <= 0 {
		tplData.CurrentPage = 1
		tplData.MaxPage = 1
		tplData.ChildItems = dirItems
	} else {
		tplData.MaxPage = int(math.Ceil(float64(len(dirItems)) / float64(tplData.PageSize)))

		switch {
		case pageNumber <= 0:
			tplData.CurrentPage = 1
		case pageNumber > tplData.MaxPage:
			tplData.CurrentPage = tplData.MaxPage
		default:
			tplData.CurrentPage = pageNumber
		}

		startIdx := (tplData.CurrentPage - 1) * tplData.PageSize
		endIdx := (tplData.CurrentPage * tplData.PageSize)
		if nFiles := len(dirItems); endIdx > nFiles {
			endIdx = nFiles
		}

		tplData.ChildItems = dirItems[startIdx:endIdx]
	}

	// Create child URLs
	childURLs := []string{}

	for _, child := range tplData.ChildItems {
		childURLs = append(childURLs, strings.TrimPrefix(child.URLPath, "/"))
	}

	for _, tag := range tplData.ChildTags {
		childURLs = append(childURLs, strings.TrimPrefix(tag.URLPath, "/"))
	}

	if tplData.MaxPage > 1 {
		for i := 1; i <= tplData.MaxPage; i++ {
			childURLs = append(childURLs, path.Join(cleanURLPath, strconv.Itoa(i)))
		}
	}

	// Render HTML
	theme := meta.Theme
	templateName := meta.DirTemplate
	if templateName == "" {
		templateName = "directory"
	}

	return childURLs, wk.renderHTML(w, tplData, theme, templateName)
}
