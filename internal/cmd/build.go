package cmd

import (
	"os"

	"github.com/go-boom/boom/internal/builder"
	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [path]",
		Short: "Build the static site",
		Args:  cobra.MaximumNArgs(1),
		Run:   buildHandler,
	}

	cmd.Flags().StringP("output", "o", "public", "path to output directory")
	return cmd
}

func buildHandler(cmd *cobra.Command, args []string) {
	rootDir := "/home/radhi/Public/new-blog"
	mdPath := "blog/#golang"
	worker, err := builder.NewWorker(rootDir, true)
	panicError(err)

	dst, _ := os.Create("Result.html")
	defer dst.Close()

	err = worker.Build(mdPath, dst)
	panicError(err)
}
