package cmd

import "github.com/spf13/cobra"

func buildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the static site",
		Args:  cobra.NoArgs,
	}

	cmd.Flags().StringP("output", "o", "public", "path to output directory")
	return cmd
}
