package cmd

import (
	"github.com/spf13/cobra"
)

// BoomCmd creates new command for boom
func BoomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "boom",
		Short: "Simple static site generator",
	}

	cmd.AddCommand(newCmd(), serveCmd(), buildCmd())
	return cmd
}
