package cmd

import (
	"github.com/RadhiFadlillah/boom/internal/webserver"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server [root-path]",
		Short:   "Run webserver for the site",
		Aliases: []string{"serve"},
		Args:    cobra.MaximumNArgs(1),
		Run:     serveHandler,
	}

	cmd.Flags().IntP("port", "p", 8080, "Port that used by webserver")
	return cmd
}

func serveHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	port, _ := cmd.Flags().GetInt("port")

	// Parse args
	rootDir := "."
	if len(args) > 0 {
		rootDir = args[0]
	}

	// Start server
	logrus.Printf("Serve boom in :%d\n", port)
	err := webserver.Start(rootDir, port)
	panicError(err)
}
