package build

import (
	"bytes"
	"strconv"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func isNumber(str string) (bool, int) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return false, 0
	}
	return true, num
}

func convertMarkdownToHTML(bt []byte) ([]byte, error) {
	highlighter := highlighting.NewHighlighting(
		highlighting.WithFormatOptions(
			chromahtml.WithClasses(true),
		),
	)

	md := goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
			extension.Footnote,
			emoji.Emoji,
			highlighter,
			mathjax.MathJax,
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	buf := bytes.NewBuffer(nil)
	err := md.Convert(bt, buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
