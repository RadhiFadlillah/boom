package build

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	fp "path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/RadhiFadlillah/boom/internal/fileutils"
	"github.com/RadhiFadlillah/boom/internal/model"
)

// buildTagFiles builds tag files list for specified URL path.
func (wk *Worker) buildTagFiles(urlPath string, w io.Writer) ([]string, error) {
	// Fetch page number and tag name from URL
	tagName := ""
	pageNumber := 1
	cleanURLPath := urlPath

	urlPathBase := path.Base(cleanURLPath)
	if isNum, number := isNumber(urlPathBase); isNum {
		pageNumber = number
		cleanURLPath = path.Dir(cleanURLPath)
	}

	urlPathBase = path.Base(cleanURLPath)
	if strings.HasPrefix(urlPathBase, "tag-") {
		tagName = strings.TrimPrefix(urlPathBase, "tag-")
		cleanURLPath = path.Dir(cleanURLPath)
	}

	// Now since the URL path clean from tag name and page number,
	// we can generate path to _index.md file from it
	dirPath := fp.Join(wk.ContentDir, cleanURLPath)
	indexMdPath := fp.Join(wk.ContentDir, cleanURLPath, "_index.md")
	if !fileutils.IsDir(dirPath) {
		return nil, fmt.Errorf("%s is not part of site content", urlPath)
	}

	// Parse metadata
	meta, _, err := wk.parsePath(indexMdPath)
	if err != nil {
		return nil, err
	}

	// Create template data
	tplData := model.TagFilesData{
		URLPath:   path.Join("/", urlPath),
		ActiveTag: tagName,
		Title:     meta.Title,
		PageSize:  meta.Pagination,
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

	tplData.PathTrails = append(tplData.PathTrails,
		model.ContentPath{
			URLPath: path.Join("/", cleanURLPath),
			Title:   meta.Title,
		},
		model.ContentPath{
			URLPath: path.Join("/", cleanURLPath, "tag-"+tagName),
			Title:   "#" + tagName,
		},
	)

	// Fetch files that uses our active tag
	files := []model.ContentPath{}
	fnWalk := func(fPath string, info os.FileInfo, err error) error {
		// We look for markdown file
		if info.IsDir() || fp.Ext(fPath) != ".md" || fp.Base(fPath) == "_index.md" {
			return nil
		}

		// Parse file
		fileMeta, _, err := wk.parsePath(fPath)
		if err != nil {
			return err
		}

		if fileMeta.Draft && !wk.buildDraft {
			return nil
		}

		// Make sure this file uses active tag
		useTag := false
		for _, tag := range fileMeta.Tags {
			if tag == tplData.ActiveTag {
				useTag = true
				break
			}
		}

		if !useTag {
			return nil
		}

		// Generate URL path
		fPath = strings.TrimSuffix(fPath, ".md")
		fileURLPath, err := fp.Rel(wk.ContentDir, fPath)
		if err != nil {
			return err
		}

		// Add it to list of file
		fileTime := fileMeta.UpdateTime
		if fileTime.IsZero() {
			fileTime = fileMeta.CreateTime
		}

		files = append(files, model.ContentPath{
			Title:      fileMeta.Title,
			URLPath:    path.Join("/", fileURLPath),
			UpdateTime: fileTime,
		})
		return nil
	}

	err = fp.Walk(dirPath, fnWalk)
	if err != nil {
		return nil, err
	}

	// Sort files
	sort.Slice(files, func(a, b int) bool {
		timeA := files[a].UpdateTime
		timeB := files[b].UpdateTime
		if !timeA.Equal(timeB) {
			return timeA.After(timeB)
		}

		titleA := files[a].Title
		titleB := files[b].Title
		return strings.ToLower(titleA) < strings.ToLower(titleB)
	})

	// Calculate pagination stuffs
	if tplData.PageSize <= 0 {
		tplData.CurrentPage = 1
		tplData.MaxPage = 1
		tplData.Files = files
	} else {
		tplData.MaxPage = int(math.Ceil(float64(len(files)) / float64(tplData.PageSize)))

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
		if nFiles := len(files); endIdx > nFiles {
			endIdx = nFiles
		}

		tplData.Files = files[startIdx:endIdx]
	}

	// Create child URLs
	childURLs := []string{}
	if tplData.MaxPage > 1 {
		for i := 1; i <= tplData.MaxPage; i++ {
			pageURL := path.Join(cleanURLPath, "tag-"+tagName, strconv.Itoa(i))
			childURLs = append(childURLs, pageURL)
		}
	}

	// Render HTML
	theme := meta.Theme
	templateName := meta.TagFilesTemplate
	if templateName == "" {
		templateName = "tagfiles"
	}

	return childURLs, wk.renderHTML(w, tplData, theme, templateName)
}
