package cmd

import "github.com/spf13/cobra"

func DocCmd() (command *cobra.Command) {
	var fileFormat string
	var filePath string
	command = &cobra.Command{
		Use: "doc",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	flags := command.Flags()
	flags.StringVarP(&fileFormat, "format", "f", "md", "doc file format")
	flags.StringVarP(&filePath, "path", "p", "./", "doc out file dir")
	return
}
