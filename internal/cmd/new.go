package cmd

import "github.com/spf13/cobra"

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new site or metadata",
	}

	cmd.AddCommand(newSiteCmd(), newMetaCmd())
	return cmd
}
