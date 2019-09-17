package summary

import (
	"fmt"
	"github.com/haozzzzzzzz/go-tool/api/info"
	"github.com/spf13/cobra"
)

func CommandVersion() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "tool version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(info.Info())
		},
	}

	return cmd
}
