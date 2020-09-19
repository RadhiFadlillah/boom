package build

import (
	"html/template"
	"path"
	fp "path/filepath"
	"strconv"
)

func (wk Worker) funcMap() template.FuncMap {
	return template.FuncMap{
		"add":            mathAdd,
		"sub":            mathSub,
		"paginationLink": wk.paginationLink,
	}
}

func mathAdd(a, b int) int {
	return a + b
}

func mathSub(a, b int) int {
	return a - b
}

func (wk Worker) paginationLink(currentPath string, pageNumber int) string {
	for {
		isNum, _ := isNumber(path.Base(currentPath))
		fPath := fp.Join(wk.ContentDir, currentPath+".md")
		if !isNum && !isFile(fPath) {
			break
		}

		currentPath = path.Dir(currentPath)
	}

	strNumber := strconv.Itoa(pageNumber)
	return path.Join(currentPath, strNumber)
}
