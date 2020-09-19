package build

import (
	"html/template"
	"path"
	"strconv"
)

var funcMap = template.FuncMap{
	"add":            mathAdd,
	"sub":            mathSub,
	"paginationLink": paginationLink,
}

func mathAdd(a, b int) int {
	return a + b
}

func mathSub(a, b int) int {
	return a - b
}

func paginationLink(currentPath string, pageNumber int) string {
	strNumber := strconv.Itoa(pageNumber)
	currentDir := currentPath
	if isNum, _ := isNumber(path.Base(currentPath)); isNum {
		currentDir = path.Dir(currentPath)
	}

	return path.Join(currentDir, strNumber)
}
