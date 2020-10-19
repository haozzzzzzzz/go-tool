package cmd

import (
	"fmt"
	"github.com/haozzzzzzzz/go-tool/ws/info"
	"github.com/spf13/cobra"
)

func CommandVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "tool version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(info.Info())
		},
	}
}
