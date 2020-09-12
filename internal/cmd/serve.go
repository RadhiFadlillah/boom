package cmd

import (
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Short:   "Run webserver for the site",
		Aliases: []string{"serve"},
		Args:    cobra.NoArgs,
		Run:     serveHandler,
	}

	cmd.Flags().IntP("port", "p", 8080, "Port that used by webserver")
	return cmd
}

func serveHandler(cmd *cobra.Command, args []string) {
}
