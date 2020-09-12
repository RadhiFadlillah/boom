package cmd

import (
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var (
	cBold  = color.Bold
	cError = color.New(color.FgRed, color.OpBold)
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
