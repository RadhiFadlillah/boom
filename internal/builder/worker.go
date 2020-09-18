package builder

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"regexp"
	"strings"

	"github.com/go-boom/boom/internal/model"
	"github.com/pelletier/go-toml"
)

var (
	rxTagURL = regexp.MustCompile(`(?i)#([^\s#/]+)(?:\/(\d+))?$`)
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

// Build builds HTML for specified URL path.
// There are two possible URL path combination :
// 1. It's pointed directly to content, e.g. /blog/awesome or /blog/awesome/1
// 2. It's URL for tag list, e.g. /blog/awesome/#cat or /blog/awesome/#cat/2
func (wk *Worker) Build(urlPath string, w io.Writer) error {
	// Trim trailing slash and hash from URL path
	for {
		urlPathLength := len(urlPath)
		urlPath = strings.Trim(urlPath, "/")
		urlPath = strings.TrimSuffix(urlPath, "#")

		if len(urlPath) == urlPathLength {
			break
		}
	}

	// Build page depending on URL path
	if rxTagURL.MatchString(urlPath) {
		return wk.buildTagList(urlPath, w)
	}

	return wk.buildPage(urlPath, w)
}

// createTemplate creates HTML template from specified theme and template name.
func (wk *Worker) renderHTML(w io.Writer, data interface{}, themeName string, templateName string) error {
	// Find theme dir
	themeDir := fp.Join(wk.rootDir, "themes", themeName)

	// If theme name not specified, use the first dir found
	if themeName == "" && isDir(themeDir) {
		dirItems, err := ioutil.ReadDir(themeDir)
		if err != nil {
			return err
		}

		for _, item := range dirItems {
			if item.IsDir() {
				themeDir = fp.Join(themeDir, item.Name())
				break
			}
		}
	}

	// Get all HTML files in theme dir
	dirItems, err := ioutil.ReadDir(themeDir)
	if err != nil {
		return err
	}

	// Separate base template and the others
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

	bt, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(bt))

	return tpl.Execute(w, data)
}

// parsePath parse markdown file in specified path. It's like `parseMarkdown`
// method, but here we also do caching and look up to parent's metadata to fill
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
			meta.TagListTemplate != "" &&
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

		if meta.TagListTemplate == "" {
			meta.TagListTemplate = parentMeta.TagListTemplate
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

// parseMarkdown parse markdown file in specified path. It will splits between
// metadata and content.
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
