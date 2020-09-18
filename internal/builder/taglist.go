package builder

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	fp "path/filepath"
	"sort"
	"strings"

	"github.com/go-boom/boom/internal/model"
)

// buildTagList builds tag list page for specified URL path.
func (wk *Worker) buildTagList(urlPath string, w io.Writer) error {
	// Fetch page number and tag name from URL
	tagName := ""
	pageNumber := 1
	urlPathSegments := strings.Split(urlPath, "/")

	urlPathBase := path.Base(urlPath)
	if isNum, number := isNumber(urlPathBase); isNum {
		pageNumber = number
		urlPathSegments = urlPathSegments[:len(urlPathSegments)-1]
	}

	urlPathBase = path.Base(urlPath)
	if strings.HasPrefix(urlPathBase, "#") {
		tagName = strings.TrimPrefix(urlPathBase, "#")
		urlPathSegments = urlPathSegments[:len(urlPathSegments)-1]
	}

	// Now since the URL path clean from tag name and page number,
	// we can generate path to _index.md file from it
	cleanURLPath := path.Join(urlPathSegments...)
	dirPath := fp.Join(wk.contentDir, cleanURLPath)
	dirIndexMdPath := fp.Join(dirPath, "_index.md")
	if !isDir(dirPath) {
		return fmt.Errorf("%s is not part of site content", urlPath)
	}

	// Parse the markdown path
	dirMeta, _, err := wk.parsePath(dirIndexMdPath)
	if err != nil {
		return err
	}

	// Create template data
	tplData := model.TagListTemplate{
		URLPath:   urlPath,
		ActiveTag: tagName,

		DirTitle: dirMeta.Title,
		PageSize: dirMeta.Pagination,
	}

	// Create path trails
	tplData.PathTrails = []model.ContentPath{{
		URLPath: path.Join(cleanURLPath, "#"+tagName),
		Title:   "#" + tagName,
	}}

	for parentPath := cleanURLPath; parentPath != "."; parentPath = path.Dir(parentPath) {
		parentFilePath := fp.Join(wk.contentDir, parentPath)
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

	// Fetch files that uses our active tag
	files := []model.ContentPath{}
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
		path = strings.TrimSuffix(path, ".md")
		fileURLPath, err := fp.Rel(wk.contentDir, path)
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
			URLPath:    fileURLPath,
			UpdateTime: fileTime,
		})
		return nil
	}

	err = fp.Walk(dirPath, fnWalk)
	if err != nil {
		return err
	}

	// Sort files
	sort.Slice(files, func(a, b int) bool {
		timeA := files[a].UpdateTime
		timeB := files[b].UpdateTime
		return timeA.After(timeB)
	})

	for _, f := range files {
		fmt.Println(f)
	}

	// Calculate pagination stuffs, return early whenever possible
	theme := dirMeta.Theme
	templateName := dirMeta.TagListTemplate
	if templateName == "" {
		templateName = "taglist"
	}

	// Handle case when pagination not used
	if tplData.PageSize <= 0 {
		tplData.CurrentPage = 1
		tplData.MaxPage = 1
		tplData.Files = files
		return wk.renderHTML(w, tplData, theme, templateName)
	}

	// Calculate max page
	tplData.MaxPage = int(math.Ceil(float64(len(files)) / float64(tplData.PageSize)))

	// Save page number to template data
	switch {
	case pageNumber <= 0:
		tplData.CurrentPage = 1
	case pageNumber > tplData.MaxPage:
		tplData.CurrentPage = tplData.MaxPage
	default:
		tplData.CurrentPage = pageNumber
	}

	// Slice files only for this page
	startIdx := (tplData.CurrentPage - 1) * tplData.PageSize
	endIdx := (tplData.CurrentPage * tplData.PageSize)
	if nFiles := len(files); endIdx > nFiles {
		endIdx = nFiles
	}

	tplData.Files = files[startIdx:endIdx]

	return wk.renderHTML(w, tplData, theme, templateName)
}
