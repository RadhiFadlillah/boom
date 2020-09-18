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
	"strings"

	"github.com/go-boom/boom/internal/model"
)

// buildPage builds content page for specified URL path.
func (wk *Worker) buildPage(urlPath string, w io.Writer) error {
	// Fetch page number from URL
	pageNumber := -1

	urlPathBase := path.Base(urlPath)
	urlPathSegments := strings.Split(urlPath, "/")
	if isNum, number := isNumber(urlPathBase); isNum {
		pageNumber = number
		urlPathSegments = urlPathSegments[:len(urlPathSegments)-1]
	}

	// Now since the URL path clean from page number, we can generate
	// file path to process.
	cleanURLPath := path.Join(urlPathSegments...)
	mdFilePath := fp.Join(wk.ContentDir, cleanURLPath)
	dirIndexMdPath := ""

	switch {
	case isDir(mdFilePath):
		dirIndexMdPath = fp.Join(mdFilePath, "_index.md")
		mdFilePath = dirIndexMdPath

	case isFile(mdFilePath + ".md"):
		dirIndexMdPath = fp.Join(fp.Dir(mdFilePath), "_index.md")
		mdFilePath = mdFilePath + ".md"

	default:
		return fmt.Errorf("%s is not part of site content", urlPath)
	}

	// Parse file and dir
	fileMeta, fileContent, err := wk.parsePath(mdFilePath)
	if err != nil {
		return err
	}

	dirMeta, _, err := wk.parsePath(dirIndexMdPath)
	if err != nil {
		return err
	}

	// Create template data
	tplData := model.PageTemplate{
		URLPath: urlPath,

		DirTitle: dirMeta.Title,
		PageSize: dirMeta.Pagination,

		Title:       fileMeta.Title,
		Description: fileMeta.Description,
		Author:      fileMeta.Author,
		CreateTime:  fileMeta.CreateTime,
		UpdateTime:  fileMeta.UpdateTime,
		Content:     fileContent,
	}

	// Create path trails
	tplData.PathTrails = []model.ContentPath{{
		URLPath: urlPath,
		Title:   tplData.Title,
		IsDir:   dirIndexMdPath == mdFilePath,
	}}

	for parentPath := path.Dir(urlPath); parentPath != "."; parentPath = path.Dir(parentPath) {
		parentFilePath := fp.Join(wk.ContentDir, parentPath)
		parentMeta, _, err := wk.parsePath(parentFilePath)
		if err != nil {
			return err
		}

		tplData.PathTrails = append([]model.ContentPath{{
			URLPath: parentPath,
			Title:   parentMeta.Title,
			IsDir:   true,
		}}, tplData.PathTrails...)
	}

	// Fetch dir items
	dirPath := fp.Dir(dirIndexMdPath)
	dirURLPath := urlPath
	if dirIndexMdPath != mdFilePath {
		dirURLPath = path.Dir(urlPath)
	}

	items, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	subDirs := []model.ContentPath{}
	subFiles := []model.ContentPath{}
	for _, item := range items {
		itemName := item.Name()
		itemExt := fp.Ext(itemName)
		itemPath := fp.Join(dirPath, itemName)
		itemURLPath := path.Join(dirURLPath, strings.TrimSuffix(itemName, itemExt))

		itemMeta, _, err := wk.parsePath(itemPath)
		if err != nil {
			return err
		}

		if item.IsDir() && !dirIsEmpty(itemPath) {
			subDirs = append(subDirs, model.ContentPath{
				URLPath: itemURLPath,
				Title:   itemMeta.Title,
				IsDir:   true,
			})
			continue
		}

		if !item.IsDir() && itemExt == ".md" && itemName != "_index.md" {
			itemTime := itemMeta.UpdateTime
			if itemTime.IsZero() {
				itemTime = itemMeta.CreateTime
			}

			subFiles = append(subFiles, model.ContentPath{
				URLPath:    itemURLPath,
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
		return err
	}

	// Sort tags
	dirTags := []model.TagPath{}
	for tag, count := range mapDirTags {
		dirTags = append(dirTags, model.TagPath{
			URLPath: path.Join(dirURLPath, "#"+tag),
			Name:    tag,
			Count:   count,
		})
	}

	sort.Slice(dirTags, func(a, b int) bool {
		nameA := dirTags[a].Name
		nameB := dirTags[b].Name
		return strings.ToLower(nameA) < strings.ToLower(nameB)
	})

	tplData.DirTags = dirTags

	// Calculate pagination stuffs, return early whenever possible
	theme := fileMeta.Theme
	templateName := fileMeta.Template
	if templateName == "" {
		templateName = "default"
	}

	sliceDirItems := func(pageNumber int) []model.ContentPath {
		startIdx := (pageNumber - 1) * tplData.PageSize
		endIdx := (pageNumber * tplData.PageSize)
		if nItems := len(dirItems); endIdx > nItems {
			endIdx = nItems
		}

		return dirItems[startIdx:endIdx]
	}

	// Handle case when pagination not used
	if tplData.PageSize <= 0 {
		tplData.CurrentPage = 1
		tplData.MaxPage = 1
		tplData.DirItems = dirItems
		return wk.renderHTML(w, tplData, theme, templateName)
	}

	// Calculate max page
	tplData.MaxPage = int(math.Ceil(float64(len(dirItems)) / float64(tplData.PageSize)))

	// Handle case when we building _index.md file
	if dirIndexMdPath == mdFilePath {
		switch {
		case pageNumber <= 0:
			tplData.CurrentPage = 1
		case pageNumber > tplData.MaxPage:
			tplData.CurrentPage = tplData.MaxPage
		default:
			tplData.CurrentPage = pageNumber
		}

		tplData.DirItems = sliceDirItems(tplData.CurrentPage)
		return wk.renderHTML(w, tplData, theme, templateName)
	}

	// At this point, we are building plain markdown page
	mdFileIdx := 0
	for i, item := range dirItems {
		if !item.IsDir && item.URLPath == urlPath {
			mdFileIdx = i
			break
		}
	}

	tplData.CurrentPage = int(math.Ceil(float64(mdFileIdx+1) / float64(tplData.PageSize)))
	tplData.DirItems = sliceDirItems(tplData.CurrentPage)

	// Create file tags
	sort.Strings(fileMeta.Tags)
	for _, tag := range fileMeta.Tags {
		tplData.Tags = append(tplData.Tags, model.TagPath{
			URLPath: path.Join(path.Dir(urlPath), "#"+tag),
			Name:    tag,
		})
	}

	// Get sibling file
	if mdFileIdx > 0 {
		if nextFile := dirItems[mdFileIdx-1]; !nextFile.IsDir {
			tplData.NextFile = nextFile
		}
	}

	if mdFileIdx < len(dirItems)-1 {
		if prevFile := dirItems[mdFileIdx+1]; !prevFile.IsDir {
			tplData.PrevFile = prevFile
		}
	}

	return wk.renderHTML(w, tplData, theme, templateName)
}
