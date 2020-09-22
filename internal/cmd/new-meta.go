package cmd

import (
	"bufio"
	"bytes"
	"io"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/go-boom/boom/internal/model"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

func newMetaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta [path]",
		Short: "Create a metadata for file at specified path",
		Args:  cobra.ExactArgs(1),
		Run:   newMetaHandler,
	}

	return cmd
}

func newMetaHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	fPath := args[0]

	// Make sure file is markdown
	if isDir(fPath) {
		fPath = fp.Join(fPath, "_index.md")
	} else {
		fName := fp.Base(fPath)
		if fp.Ext(fName) != ".md" {
			fName += ".md"
		}
		fPath = fp.Join(fp.Dir(fPath), fName)
	}

	// Open file
	f, err := os.OpenFile(fPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	panicError(err)
	defer f.Close()

	// Separate metadata and content of file
	separatorCount := 0
	metaBuffer := bytes.NewBuffer(nil)
	contentBuffer := bytes.NewBuffer(nil)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "+++" {
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
	var meta model.Metadata
	toml.NewDecoder(metaBuffer).Decode(&meta)

	// Ask metadata from user
	scanner = bufio.NewScanner(os.Stdin)

	if meta.Title == "" {
		cBold.Print("Title : ")
		scanner.Scan()
		meta.Title = strings.TrimSpace(scanner.Text())
	}

	if meta.Description == "" {
		cBold.Print("Description : ")
		scanner.Scan()
		meta.Description = strings.TrimSpace(scanner.Text())
	}

	if meta.Author == "" {
		cBold.Print("Author : ")
		scanner.Scan()
		meta.Author = strings.TrimSpace(scanner.Text())
	}

	currentTime := time.Now()
	if meta.CreateTime.IsZero() {
		meta.CreateTime = currentTime
		meta.UpdateTime = currentTime
	}

	if meta.UpdateTime.IsZero() {
		meta.UpdateTime = meta.CreateTime
	}

	if len(meta.Tags) == 0 {
		cBold.Print("Tags : ")
		scanner.Scan()
		strTags := strings.TrimSpace(scanner.Text())

		if strTags != "" {
			tags := strings.Split(strTags, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			meta.Tags = tags
		}
	}

	cBold.Print("Draft : ")
	scanner.Scan()
	meta.Draft = strings.TrimSpace(scanner.Text()) == "1"

	// Encode metadata
	bt, err := toml.Marshal(meta)
	panicError(err, "Failed to create metadata:")

	// Merge meta and content
	buf := bytes.NewBuffer(nil)
	buf.WriteString("+++\n")
	buf.Write(bt)
	buf.WriteString("+++\n")
	buf.Write(contentBuffer.Bytes())

	// Truncate file
	err = f.Truncate(0)
	panicError(err)

	_, err = f.Seek(0, 0)
	panicError(err)

	_, err = io.Copy(f, buf)
	panicError(err)
}
