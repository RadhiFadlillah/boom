package cmd

import (
	"bufio"
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/go-boom/boom/internal/model"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [path]",
		Short: "Create a new site at specified path",
		Args:  cobra.ExactArgs(1),
		Run:   newHandler,
	}

	cmd.Flags().StringP("title", "t", "", "title of the website")
	cmd.Flags().StringP("owner", "o", "", "owner of the website")
	cmd.Flags().BoolP("force", "f", false, "force init inside non-empty directory")
	cmd.AddCommand()
	return cmd
}

func newHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	rootDir := args[0]
	title, _ := cmd.Flags().GetString("title")
	owner, _ := cmd.Flags().GetString("owner")
	isForced, _ := cmd.Flags().GetBool("force")

	title = strings.TrimSpace(title)
	owner = strings.TrimSpace(owner)

	// Make sure target directory exists
	os.MkdirAll(rootDir, os.ModePerm)

	// Make sure target dir is empty
	if !dirEmpty(rootDir) && !isForced {
		cError.Printf("Directory %s already exists and not empty\n", rootDir)
		return
	}

	// Get sites metadata from user
	scanner := bufio.NewScanner(os.Stdin)
	if title == "" {
		cBold.Print("Website title : ")
		scanner.Scan()
		title = strings.TrimSpace(scanner.Text())

		if title == "" {
			cError.Println("Website title must not empty")
			return
		}
	}

	if owner == "" {
		cBold.Print("Website owner : ")
		scanner.Scan()
		owner = strings.TrimSpace(scanner.Text())
	}

	// Create directories
	os.MkdirAll(fp.Join(rootDir, "themes"), os.ModePerm)
	os.MkdirAll(fp.Join(rootDir, "content"), os.ModePerm)

	// Write first page
	prefixErrIndex := "Failed to create index page:"

	indexPath := fp.Join(rootDir, "content", "_index.md")
	indexFile, err := os.Create(indexPath)
	panicError(err, prefixErrIndex)
	defer indexFile.Close()

	_, err = indexFile.WriteString("Hello World")
	panicError(err, prefixErrIndex)

	err = indexFile.Sync()
	panicError(err, prefixErrIndex)

	// Write metadata
	prefixErrMeta := "Failed to create metadata:"

	metaPath := fp.Join(rootDir, "content", "_meta.toml")
	metaFile, err := os.Create(metaPath)
	panicError(err, prefixErrMeta)
	defer metaFile.Close()

	currentTime := time.Now()
	err = toml.NewEncoder(metaFile).Encode(model.Metadata{
		Title:      title,
		Author:     owner,
		CreateTime: currentTime,
		UpdateTime: currentTime,
		Pagination: 10})
	panicError(err, prefixErrMeta)

	// Finish
	fmt.Print("Your new site is created in ")
	cBold.Println(rootDir)
}
