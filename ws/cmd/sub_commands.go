package cmd

import "github.com/spf13/cobra"

func SubCommands(parent *cobra.Command) {
	parent.AddCommand(DocCmd())
	parent.AddCommand(CommandVersion())
}
