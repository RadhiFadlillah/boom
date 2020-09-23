package cmd

import (
	"bufio"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/go-boom/boom/internal/build"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [path]",
		Short: "Build the static site",
		Args:  cobra.MaximumNArgs(1),
		Run:   buildHandler,
	}

	cmd.Flags().StringP("output", "o", "", "path to output directory")
	return cmd
}

func buildHandler(cmd *cobra.Command, args []string) {
	// Parse args
	rootDir := "."
	if len(args) > 0 {
		rootDir = args[0]
	}

	rootDir, err := fp.Abs(rootDir)
	panicError(err)

	// Parse flags
	outputDir, _ := cmd.Flags().GetString("output")
	if outputDir == "" {
		outputDir = fp.Join(rootDir, "public")
	}

	// Clean output dir, but keep CNAME file if it exists
	logrus.Println("cleaning output dir")
	if !isDir(outputDir) {
		os.MkdirAll(outputDir, os.ModePerm)
	} else {
		items, err := ioutil.ReadDir(outputDir)
		panicError(err)

		for _, item := range items {
			itemName := item.Name()
			if !item.IsDir() && strings.ToLower(itemName) == "cname" {
				continue
			}

			itemPath := fp.Join(outputDir, itemName)
			os.RemoveAll(itemPath)
		}
	}

	// Copy assets
	assetsDir := fp.Join(rootDir, "assets")
	if isDir(assetsDir) {
		logrus.Println("copying assets")
		dstDir := fp.Join(outputDir, "assets")
		err := copyDir(assetsDir, dstDir, nil)
		panicError(err)
	}

	// Copy themes
	themesDir := fp.Join(rootDir, "themes")
	if isDir(themesDir) {
		logrus.Println("copying themes")

		// Create method for parsing .boomignore file
		parseBoomignore := func(themeDir, boomignorePath string) (map[string]struct{}, error) {
			boomignore, err := os.Open(boomignorePath)
			if err != nil {
				return nil, err
			}
			defer boomignore.Close()

			excludedPaths := make(map[string]struct{})
			scanner := bufio.NewScanner(boomignore)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				currentPath := fp.Clean(fp.Join(themeDir, line))
				excludedPaths[currentPath] = struct{}{}
			}

			return excludedPaths, nil
		}

		// Get list of items in themes dir
		themesDirItems, err := ioutil.ReadDir(themesDir)
		panicError(err)

		for _, item := range themesDirItems {
			if !item.IsDir() {
				continue
			}

			// Get path to .boomignore
			itemName := item.Name()
			themeDir := fp.Join(themesDir, itemName)
			boomignorePath := fp.Join(themeDir, ".boomignore")

			// Parse .boomignore file if necessary
			excludedPaths := make(map[string]struct{})
			if isFile(boomignorePath) {
				excludedPaths, err = parseBoomignore(themeDir, boomignorePath)
				panicError(err)
			}

			// Make sure .git and node_modules excluded
			excludedPaths[fp.Join(themeDir, ".git")] = struct{}{}
			excludedPaths[fp.Join(themeDir, "node_modules")] = struct{}{}

			// Copy theme
			dstDir := fp.Join(outputDir, "themes", itemName)
			copyDir(themeDir, dstDir, excludedPaths)
		}
	}

	// Build site content
	cfg := build.Config{
		EnableCache:  true,
		BuildDraft:   false,
		MinifyOutput: true,
	}

	processedURLs := make(map[string]struct{})
	wk, err := build.NewWorker(rootDir, cfg)
	panicError(err)

	var buildFunc func(string) error
	buildFunc = func(urlPath string) error {
		// Dont build reserved directory
		if urlPath == "assets" || urlPath == "themes" {
			return nil
		}

		// If this URL already build, stop
		if _, exist := processedURLs[urlPath]; exist {
			return nil
		}
		logrus.Printf("building /%s\n", urlPath)

		// Create destination file
		dstPath := fp.Join(outputDir, urlPath, "index.html")
		os.MkdirAll(fp.Dir(dstPath), os.ModePerm)

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// Build page
		childURLs, err := wk.Build(urlPath, dstFile)
		if err != nil {
			os.Remove(dstPath)
			if err != build.ErrDraftFile {
				return err
			}
		}

		// Save this URL as processed
		processedURLs[urlPath] = struct{}{}

		// Build each child URL
		for _, childURL := range childURLs {
			err = buildFunc(childURL)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = buildFunc("")
	panicError(err)
}
