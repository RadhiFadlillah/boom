package cmd

import (
	"os"
	fp "path/filepath"

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

	// Parse flags
	outputDir, _ := cmd.Flags().GetString("output")
	if outputDir == "" {
		outputDir = fp.Join(rootDir, "public")
	}

	// Clean output dir
	logrus.Println("cleaning output dir")
	os.RemoveAll(outputDir)
	os.MkdirAll(outputDir, os.ModePerm)

	// Copy assets
	assetsDir := fp.Join(rootDir, "assets")
	if isDir(assetsDir) {
		logrus.Println("copying assets")
		dstDir := fp.Join(outputDir, "assets")
		err := copyDir(assetsDir, dstDir)
		panicError(err)
	}

	// Copy themes
	themesDir := fp.Join(rootDir, "themes")
	if isDir(themesDir) {
		logrus.Println("copying themes")
		dstDir := fp.Join(outputDir, "themes")
		err := copyDir(themesDir, dstDir)
		panicError(err)
	}

	// Build site content
	processedURLs := make(map[string]struct{})
	wk, err := build.NewWorker(rootDir, true, false)
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
