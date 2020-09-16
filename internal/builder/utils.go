package builder

import (
	"bytes"
	"os"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func isDir(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return f.IsDir()
}

func isFile(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !f.IsDir()
}

func renderHTMLMarkdown(bt []byte) ([]byte, error) {
	highlighter := highlighting.NewHighlighting(
		highlighting.WithStyle("emacs"),
		highlighting.WithFormatOptions(
			chromahtml.WithClasses(true),
			chromahtml.WithLineNumbers(true),
			chromahtml.LinkableLineNumbers(true, ""),
			chromahtml.LineNumbersInTable(true),
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
