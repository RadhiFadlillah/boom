package builder

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"
	fp "path/filepath"
	"sort"
	"strings"

	"github.com/go-boom/boom/internal/model"
	"github.com/pelletier/go-toml"
)

// Worker is the one that build markdown into HTML file.
type Worker struct {
	rootDir      string
	contentDir   string
	cacheEnabled bool

	metaCache map[string]model.Metadata
	htmlCache map[string]template.HTML
}

// NewWorker returns a new worker. Requires root dir which point to directory
// where site lives.
func NewWorker(rootDir string, enableCache bool) (wk Worker, err error) {
	// Make sure root dir is a valid dir
	if !isDir(rootDir) {
		err = errors.New("the specified root dir is not a directory")
		return
	}

	// Validate content directory
	contentDir := fp.Join(rootDir, "content")
	if !isDir(contentDir) {
		err = errors.New("content dir doesn't exist")
		return
	}

	contentIndexPath := fp.Join(contentDir, "_index.md")
	if _, _, err = wk.parseMarkdown(contentIndexPath); err != nil {
		return
	}

	// Create a new worker
	wk = Worker{
		rootDir:      rootDir,
		contentDir:   contentDir,
		cacheEnabled: enableCache,
		metaCache:    make(map[string]model.Metadata),
		htmlCache:    make(map[string]template.HTML),
	}
	return
}

// BuildPage builds page for specified path. Path here is not filepath though,
// but rather an URL path. The build result will be written to writer.
func (wk *Worker) BuildPage(urlPath string, w io.Writer) error {
	// Trim trailing slash from URL path
	urlPath = strings.Trim(urlPath, "/")

	// Convert URL path to file path.
	// There are several possible URL path combination :
	// 1. It's pointed directly to content, e.g. /blog/awesome, which might be for :
	//    - content/blog/awesome.md
	//    - content/blog/awesome/_index.md
	// 2. It's URL for pagination, e.g. /blog/awesome/1
	// 3. TODO: It's URL for tags, e.g. /blog/awesome/#cat or /blog/awesome/#cat/2
	// activeTag := ""
	pageNumber := -1

	// Check if it's ended with page number
	urlPathBase := path.Base(urlPath)
	urlPathSegments := strings.Split(urlPath, "/")
	if isNum, number := isNumber(urlPathBase); isNum {
		pageNumber = number
		urlPathSegments = urlPathSegments[:len(urlPathSegments)-1]
	}

	// Check if it's for tag
	urlPathBase = path.Base(urlPath)
	if strings.HasPrefix(urlPathBase, "#") {
		// activeTag = strings.TrimPrefix(urlPathBase, "#")
		urlPathSegments = urlPathSegments[:len(urlPathSegments)-1]
	}

	// At this point our URL path should be in pattern 1,
	// so we can generate the filepath
	dirIndexMdPath := ""
	mdFilePath := fp.Join(urlPathSegments...)
	mdFilePath = fp.Join(wk.contentDir, mdFilePath)

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
	mapDirTags := make(map[string]int)
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

		if !item.IsDir() && itemExt == ".md" {
			for _, tag := range itemMeta.Tags {
				mapDirTags[tag]++
			}

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

	// Create dir tags
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

	// At this point we only need to calculate pagination stuffs,
	// so return early whenever possible. There are several cases
	// that we want to look for :
	// - Pagination is not used (page size <= 0)
	// - Pagination is used and we are building index page of directory
	// - Pagination is used and we are building markdown page
	theme := fileMeta.Theme
	templateName := fileMeta.Template
	sliceDirItems := func(pageNumber int) []model.ContentPath {
		itemStartIdx := (pageNumber - 1) * tplData.PageSize
		itemEndIdx := (pageNumber * tplData.PageSize)
		return dirItems[itemStartIdx:itemEndIdx]
	}

	// No pagination
	if tplData.PageSize <= 0 {
		tplData.CurrentPage = 1
		tplData.MaxPage = 1
		tplData.DirItems = dirItems
		return wk.renderHTML(w, tplData, theme, templateName)
	}

	// Building index page of directory
	tplData.MaxPage = int(math.Ceil(float64(len(tplData.DirItems)) / float64(tplData.PageSize)))

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

	// Finally, building plain markdown page
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

// createTemplate creates HTML template from specified theme and template name.
func (wk *Worker) renderHTML(w io.Writer, data interface{}, themeName string, templateName string) error {
	// Get all HTML files in theme dir
	themeDir := fp.Join(wk.rootDir, "themes", themeName)
	dirItems, err := ioutil.ReadDir(themeDir)
	if err != nil {
		return err
	}

	// Separate base template and the others
	if templateName == "" {
		templateName = "default"
	}

	baseTemplate := fp.Join(themeDir, templateName) + ".html"
	templateFiles := []string{baseTemplate}

	for _, item := range dirItems {
		name := item.Name()
		switch {
		case item.IsDir(),
			fp.Ext(name) != ".html",
			name == templateName+".html":
			continue
		}

		templateFiles = append(templateFiles, fp.Join(themeDir, name))
	}

	// Create template
	tpl, err := template.ParseFiles(templateFiles...)
	if err != nil {
		return err
	}

	return tpl.Execute(w, data)
}

// parsePath parse markdown file in specified path.
// It's like `parseMarkdown` method, but here we also do
// caching and look up to parent's metadata to fill
// missing metadata in current path.
func (wk *Worker) parsePath(path string) (meta model.Metadata, htmlContent template.HTML, err error) {
	// Check if this path already cached
	if wk.cacheEnabled {
		var metaExist, htmlExist bool

		meta, metaExist = wk.metaCache[path]
		htmlContent, htmlExist = wk.htmlCache[path]
		if metaExist && htmlExist {
			return
		}
	}

	// Capture metadata from markdown file.
	// If this path points to markdown file, just process it.
	// If this path points to directory, look for _index.md file.
	indexMd := fp.Join(path, "_index.md")
	switch {
	case isFile(path) && fp.Ext(path) == ".md":
		meta, htmlContent, err = wk.parseMarkdown(path)
	case isDir(path) && isFile(indexMd):
		meta, htmlContent, err = wk.parseMarkdown(indexMd)
	}

	// Sometimes user might not fill nor create the metadata.
	// In this case, looks for parent's metadata.
	metaIsReady := func() bool {
		return meta.Title != "" &&
			meta.Theme != "" &&
			meta.Template != "" &&
			meta.Pagination != 0
	}

	for parent := fp.Dir(path); parent != wk.rootDir; parent = fp.Dir(parent) {
		// If all important meta is ready, stop
		if metaIsReady() {
			break
		}

		// Get parent metadata
		parentIndex := fp.Join(parent, "_index.md")
		if !isFile(parentIndex) {
			continue
		}

		parentMeta, _, parentErr := wk.parseMarkdown(parentIndex)
		if parentErr != nil {
			err = parentErr
			return
		}

		// Fill metadata
		if meta.Title == "" {
			meta.Title = parentMeta.Title
		}

		if meta.Theme == "" {
			meta.Theme = parentMeta.Theme
		}

		if meta.Template == "" {
			meta.Template = parentMeta.ChildTemplate
		}

		if meta.Pagination == 0 {
			meta.Pagination = parentMeta.Pagination
		}
	}

	// Save to cache
	if wk.cacheEnabled {
		wk.metaCache[path] = meta
		wk.htmlCache[path] = htmlContent
	}

	return
}

// parseMarkdown parse markdown file in specified path.
// It will splits between metadata and content.
func (wk *Worker) parseMarkdown(mdPath string) (meta model.Metadata, htmlContent template.HTML, err error) {
	// Open file
	f, err := os.Open(mdPath)
	if err != nil {
		return
	}
	defer f.Close()

	// Separate metadata and content of file
	separatorCount := 0
	metaBuffer := bytes.NewBuffer(nil)
	contentBuffer := bytes.NewBuffer(nil)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "+++" {
			separatorCount++
			continue
		}

		if separatorCount == 1 {
			metaBuffer.WriteString(line)
			metaBuffer.WriteString("\n")
		} else {
			contentBuffer.WriteString(line)
			contentBuffer.WriteString("\n")
		}
	}

	// Parse metadata
	err = toml.NewDecoder(metaBuffer).Decode(&meta)
	if err != nil {
		return
	}

	// Parse markdown
	btHTML, err := convertMarkdownToHTML(contentBuffer.Bytes())
	if err != nil {
		return
	}

	htmlContent = template.HTML(btHTML)
	return
}
