package cmd

import (
	"bufio"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/RadhiFadlillah/boom/internal/build"
	"github.com/RadhiFadlillah/boom/internal/fileutils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [root-path]",
		Short: "Build the static site",
		Args:  cobra.MaximumNArgs(1),
		Run:   buildHandler,
	}

	cmd.Flags().StringP("output", "o", "", "path to output directory")
	return cmd
}

func buildHandler(cmd *cobra.Command, args []string) {
	// Save starting
	start := time.Now()

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

	// Clean output dir, but keep CNAME file and dot dir
	logrus.Println("cleaning output dir")
	err = cleanOutputDir(outputDir)
	panicError(err)

	// Copy assets
	logrus.Println("copying assets")
	err = copyAssets(rootDir, outputDir)
	panicError(err)

	// Copy themes
	logrus.Println("copying themes")
	err = copyThemes(rootDir, outputDir)
	panicError(err)

	// Build site content
	logrus.Println("building site content")
	err = buildContent(rootDir, outputDir)
	panicError(err)

	// Report build duration
	duration := time.Now().Sub(start).Milliseconds()
	logrus.Printf("finished after %d ms\n", duration)
}

func cleanOutputDir(outputDir string) error {
	// If output dir doesn't exist just make it
	if !fileutils.IsDir(outputDir) {
		return os.MkdirAll(outputDir, os.ModePerm)
	}

	// List all items in output dir
	items, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return err
	}

	// Remove everything except:
	// - assets dir
	// - themes dir
	// - dot dir (like .git)
	// - CNAME file
	for _, item := range items {
		itemName := item.Name()

		switch {
		case item.IsDir() && itemName == "assets",
			item.IsDir() && itemName == "themes",
			item.IsDir() && strings.HasPrefix(itemName, "."),
			!item.IsDir() && strings.ToLower(itemName) == "cname":
			continue
		}

		itemPath := fp.Join(outputDir, itemName)
		err = os.RemoveAll(itemPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyAssets(rootDir, outputDir string) error {
	// Create path to asset dirs
	srcDir := fp.Join(rootDir, "assets")
	dstDir := fp.Join(outputDir, "assets")

	// If source doesn't exist, remove destination
	if !fileutils.IsDir(srcDir) {
		return os.RemoveAll(dstDir)
	}

	// If destination doesn't exist, copy the entire source dir
	if !fileutils.IsDir(dstDir) {
		return fileutils.CopyDir(srcDir, dstDir, nil)
	}

	// List all items in src and dst
	srcItems := make(map[string]os.FileMode)
	dstItems := make(map[string]os.FileMode)

	fp.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		path, _ = fp.Rel(srcDir, path)
		srcItems[path] = info.Mode()
		return nil
	})

	fp.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		path, _ = fp.Rel(dstDir, path)
		dstItems[path] = info.Mode()
		return nil
	})

	// Remove items in dst that doesn't exist in src
	for dstItem := range dstItems {
		if _, exist := srcItems[dstItem]; exist {
			continue
		}

		err := os.RemoveAll(fp.Join(dstDir, dstItem))
		if err != nil {
			return err
		}
	}

	// Copy files from src to dst
	for srcItem, mode := range srcItems {
		// Ignore directory
		if mode.IsDir() {
			continue
		}

		// Create file path
		srcPath := fp.Join(srcDir, srcItem)
		dstPath := fp.Join(dstDir, srcItem)

		// If file is the same, continue
		if fileutils.SameFile(srcPath, dstPath) {
			continue
		}

		// If src and dst is different, copy
		err := os.RemoveAll(dstPath)
		if err != nil {
			return err
		}

		err = fileutils.CopyFile(srcPath, dstPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyThemes(rootDir, outputDir string) error {
	// Create path to themes dirs
	srcDir := fp.Join(rootDir, "themes")
	dstDir := fp.Join(outputDir, "themes")

	// If source doesn't exist, remove destination
	if !fileutils.IsDir(srcDir) {
		return os.RemoveAll(dstDir)
	}

	// Get list of excluded paths from each theme
	excludedPaths := make(map[string]struct{})
	themeList, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, theme := range themeList {
		// Valid theme must be a directory
		if !theme.IsDir() {
			continue
		}

		// If theme has boomignore file, parse it
		themeDir := fp.Join(srcDir, theme.Name())
		boomignorePath := fp.Join(themeDir, ".boomignore")

		if fileutils.IsFile(boomignorePath) {
			err = func() error {
				// Open boomignore file
				boomignore, err := os.Open(boomignorePath)
				if err != nil {
					return nil
				}
				defer boomignore.Close()

				// Read each line
				scanner := bufio.NewScanner(boomignore)
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if line == "" {
						continue
					}

					currentPath := fp.Clean(fp.Join(themeDir, line))
					excludedPaths[currentPath] = struct{}{}
				}

				return nil
			}()

			if err != nil {
				return err
			}

			// Exclude boomignore file as well
			excludedPaths[boomignorePath] = struct{}{}
		}

		// Read items in theme's root dir
		themeItems, err := ioutil.ReadDir(themeDir)
		if err != nil {
			return err
		}

		// Make sure to exclude :
		// - dot dir (like .git)
		// - node_modules dir
		// - html file since it's only used in template
		for _, item := range themeItems {
			itemName := item.Name()
			itemPath := fp.Join(themeDir, itemName)

			switch {
			case fp.Ext(itemName) == ".html",
				item.IsDir() && itemName == "node_modules",
				item.IsDir() && strings.HasPrefix(itemName, "."):
				excludedPaths[itemPath] = struct{}{}
			}
		}
	}

	// List all items in src and dst
	srcItems := make(map[string]os.FileMode)
	dstItems := make(map[string]os.FileMode)

	fp.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		_, excluded := excludedPaths[path]
		if excluded {
			if info.IsDir() {
				return fp.SkipDir
			} else {
				return nil
			}
		}

		path, _ = fp.Rel(srcDir, path)
		srcItems[path] = info.Mode()
		return nil
	})

	fp.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		path, _ = fp.Rel(dstDir, path)
		dstItems[path] = info.Mode()
		return nil
	})

	// Remove items in dst that doesn't exist in src
	for dstItem := range dstItems {
		if _, exist := srcItems[dstItem]; exist {
			continue
		}

		err := os.RemoveAll(fp.Join(dstDir, dstItem))
		if err != nil {
			return err
		}
	}

	// Copy files from src to dst
	for srcItem, mode := range srcItems {
		// Ignore directory
		if mode.IsDir() {
			continue
		}

		// Create file path
		srcPath := fp.Join(srcDir, srcItem)
		dstPath := fp.Join(dstDir, srcItem)

		// If file is the same, continue
		if fileutils.SameFile(srcPath, dstPath) {
			continue
		}

		// If src and dst is different, copy
		err := os.RemoveAll(dstPath)
		if err != nil {
			return err
		}

		err = fileutils.CopyFile(srcPath, dstPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildContent(rootDir, outputDir string) error {
	// Create worker
	cfg := build.Config{
		EnableCache:  true,
		BuildDraft:   false,
		MinifyOutput: true,
	}

	wk, err := build.NewWorker(rootDir, cfg)
	if err != nil {
		return err
	}

	// Create method for recursively build pages
	var fnBuild func(string) error
	processedURLs := make(map[string]struct{})

	fnBuild = func(urlPath string) error {
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

		// Mark this URL as already processed
		processedURLs[urlPath] = struct{}{}

		// Build each child URL
		for _, childURL := range childURLs {
			err = fnBuild(childURL)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// Build the content
	return fnBuild("")
}
