package cmd

import (
	"bufio"
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/go-boom/boom/internal/model"
	"github.com/gookit/color"
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

	cmd.Flags().BoolP("force", "f", false, "force init inside non-empty directory")
	cmd.AddCommand()
	return cmd
}

func newHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	rootDir := args[0]
	isForced, _ := cmd.Flags().GetBool("force")

	// Make sure target directory exists
	os.MkdirAll(rootDir, os.ModePerm)

	// Make sure target dir is empty
	if !dirEmpty(rootDir) && !isForced {
		color.Error.Printf("Directory %s already exists and not empty\n", rootDir)
		return
	}

	// Get sites metadata from user
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input data for new website")
	fmt.Println()

	color.Bold.Print("Website title : ")
	scanner.Scan()
	title := strings.TrimSpace(scanner.Text())

	if title == "" {
		color.Error.Println("Website title must not empty")
		return
	}

	color.Bold.Print("Website owner : ")
	scanner.Scan()
	owner := strings.TrimSpace(scanner.Text())

	// Create directories
	os.MkdirAll(fp.Join(rootDir, "theme"), os.ModePerm)
	os.MkdirAll(fp.Join(rootDir, "content"), os.ModePerm)

	// Write first page
	indexPath := fp.Join(rootDir, "content", "_index.md")
	indexFile, err := os.Create(indexPath)
	if err != nil {
		color.Error.Println("Failed to create index page:", err)
		return
	}
	defer indexFile.Close()

	_, err = indexFile.WriteString("Hello World")
	if err != nil {
		color.Error.Println("Failed to write index page:", err)
		return
	}

	err = indexFile.Sync()
	if err != nil {
		color.Error.Println("Failed to write index page:", err)
		return
	}

	// Write metadata
	metaPath := fp.Join(rootDir, "content", "_meta.toml")
	metaFile, err := os.Create(metaPath)
	if err != nil {
		color.Error.Println("Failed to create metadata file:", err)
		return
	}
	defer metaFile.Close()

	currentTime := time.Now()
	err = toml.NewEncoder(metaFile).Encode(model.Metadata{
		Title:      title,
		Author:     owner,
		CreateTime: currentTime,
		UpdateTime: currentTime,
		Pagination: 10})
	if err != nil {
		color.Error.Println("Failed to write metadata:", err)
		return
	}

	// Finish
	fmt.Println()
	fmt.Print("Congratulations! Your new site is created in ")
	color.Bold.Println(rootDir)
}
