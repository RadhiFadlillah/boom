package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"path"
	fp "path/filepath"
	"sort"
	"strings"

	"github.com/go-boom/boom/internal/model"
)

// buildFile builds file for specified URL path.
func (wk *Worker) buildFile(urlPath string, w io.Writer) error {
	// Create file path from URL
	filePath := fp.Join(wk.ContentDir, urlPath+".md")
	if !isFile(filePath) {
		return fmt.Errorf("%s is not part of site content", urlPath)
	}

	// Parse metadata
	meta, content, err := wk.parsePath(filePath)
	if err != nil {
		return err
	}

	// Create template data
	tplData := model.FileTemplate{
		URLPath:     path.Join("/", urlPath),
		Title:       meta.Title,
		Description: meta.Description,
		Author:      meta.Author,
		CreateTime:  meta.CreateTime,
		UpdateTime:  meta.UpdateTime,
		Content:     content,
	}

	// Create path trails
	trailSegments := append([]string{"."}, strings.Split(urlPath, "/")...)
	for i := 1; i < len(trailSegments); i++ {
		parentPath := strings.Join(trailSegments[:i], "/")
		parentFilePath := fp.Join(wk.ContentDir, parentPath, "_index.md")
		parentMeta, _, err := wk.parsePath(parentFilePath)
		if err != nil {
			return err
		}

		tplData.PathTrails = append(tplData.PathTrails, model.ContentPath{
			URLPath: path.Join("/", parentPath),
			Title:   parentMeta.Title,
			IsDir:   true,
		})
	}

	tplData.PathTrails = append(tplData.PathTrails,
		model.ContentPath{
			URLPath: path.Join("/", urlPath),
			Title:   meta.Title,
		},
	)

	// Fetch file tags
	sort.Strings(meta.Tags)
	dirURLPath := path.Dir(urlPath)
	for _, tag := range meta.Tags {
		tplData.Tags = append(tplData.Tags, model.TagPath{
			URLPath: path.Join("/", dirURLPath, "tag-"+tag),
			Name:    tag,
		})
	}

	// Get sibling files
	dirPath := fp.Dir(filePath)
	dirItems, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	fileIdx := -1
	dirFiles := []model.ContentPath{}
	for _, item := range dirItems {
		// Make sure it's ordinary markdown file
		itemName := item.Name()
		itemExt := fp.Ext(itemName)
		if item.IsDir() || itemExt != ".md" || itemName == "_index.md" {
			continue
		}

		// Make sure it's not draft
		itemPath := fp.Join(dirPath, itemName)
		itemMeta, _, err := wk.parsePath(itemPath)
		if err != nil {
			return err
		}

		if itemMeta.Draft {
			continue
		}

		// Add item to file list
		itemTime := itemMeta.UpdateTime
		if itemTime.IsZero() {
			itemTime = itemMeta.CreateTime
		}

		itemName = strings.TrimSuffix(itemName, itemExt)
		itemURLPath := path.Join("/", dirURLPath, itemName)
		dirFiles = append(dirFiles, model.ContentPath{
			URLPath:    itemURLPath,
			Title:      itemMeta.Title,
			UpdateTime: itemTime,
		})

		// If this item is the current file, save its index
		fmt.Println(itemURLPath, urlPath)
		if itemURLPath == tplData.URLPath {
			fileIdx = len(dirFiles) - 1
			continue
		}

		if fileIdx >= 0 && len(dirFiles) > fileIdx+1 {
			break
		}
	}

	if fileIdx >= 0 {
		if fileIdx > 0 {
			tplData.PrevFile = dirFiles[fileIdx-1]
		}

		if fileIdx < len(dirFiles)-1 {
			tplData.NextFile = dirFiles[fileIdx+1]
		}
	}

	// Render HTML
	theme := meta.Theme
	templateName := meta.Template
	if templateName == "" {
		templateName = "file"
	}

	return wk.renderHTML(w, tplData, theme, templateName)
}
