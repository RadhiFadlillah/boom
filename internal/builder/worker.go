package builder

import (
	"bufio"
	"bytes"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/go-boom/boom/internal/model"
	"github.com/pelletier/go-toml"
)

type Worker struct {
	rootDir      string
	contentDir   string
	cacheEnabled bool

	metaCache map[string]model.Metadata
	htmlCache map[string]template.HTML
}

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

func (wk *Worker) BuildPage(path string, w io.Writer) (err error) {
	// Path must be child of content dir
	relPath, err := fp.Rel(wk.contentDir, path)
	if err != nil {
		return
	}

	if strings.HasPrefix(relPath, "..") {
		err = errors.New("path is not child of content dir")
		return
	}

	// Parse page
	meta, htmlContent, err := wk.parsePath(path)
	if err != nil {
		return
	}

	// Create template data
	tplData := model.PageTemplate{
		IsDir: isDir(path),

		Title:       meta.Title,
		Description: meta.Description,
		Author:      meta.Author,
		CreateTime:  meta.CreateTime,
		UpdateTime:  meta.UpdateTime,
		Content:     htmlContent,
	}

	// Use relative path as URL path
	tplData.URLPath = relPath
	if fp.Ext(tplData.URLPath) == ".md" {
		tplData.URLPath = strings.TrimSuffix(tplData.URLPath, ".md")
	}

	// Create path trails
	tplData.PathTrails = []model.PagePath{
		{Path: tplData.URLPath, Title: tplData.Title},
	}

	for parent := fp.Dir(path); parent != wk.rootDir; parent = fp.Dir(parent) {
		parentMeta, _, parentErr := wk.parsePath(parent)
		if parentErr != nil {
			err = parentErr
			return
		}

		parentRelPath, _ := fp.Rel(wk.contentDir, parent)
		tplData.PathTrails = append([]model.PagePath{
			{Path: parentRelPath, Title: parentMeta.Title},
		}, tplData.PathTrails...)
	}

	// Create page tags
	for _, t := range meta.Tags {
		tplData.Tags = append(tplData.Tags, model.PagePath{
			Path:  fp.Join(fp.Dir(relPath), "#"+t),
			Title: "#" + t,
		})
	}

	// Open template file
	tpl, err := wk.createTemplate(meta.Theme, meta.Template)
	if err != nil {
		return
	}

	// Execute template
	err = tpl.Execute(w, tplData)
	return
}

func (wk *Worker) createTemplate(themeName string, templateName string) (*template.Template, error) {
	// Get all HTML files in theme dir
	themeDir := fp.Join(wk.rootDir, "themes", themeName)
	dirItems, err := ioutil.ReadDir(themeDir)
	if err != nil {
		return nil, err
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

	return template.ParseFiles(templateFiles...)
}

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

	// Sometimes user might not fill all metadata.
	// In this case, looks for parent's metadata.
	metaIsReady := func() bool {
		return meta.Title != "" &&
			meta.Author != "" &&
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

		if meta.Author == "" {
			meta.Author = parentMeta.Author
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
	btHTML, err := renderHTMLMarkdown(contentBuffer.Bytes())
	if err != nil {
		return
	}

	htmlContent = template.HTML(btHTML)
	return
}
