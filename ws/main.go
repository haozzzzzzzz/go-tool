package main

import (
	"github.com/haozzzzzzzz/go-tool/ws/cmd"
	"github.com/haozzzzzzzz/go-tool/ws/info"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var BuildTime = "" // 构建时间，由编译器传入

func main() {
	var err error
	info.BuildTime = BuildTime
	mainCmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				logrus.Errorf("show cmd help failed. error: %s", err)
				return
			}
		},
	}
	cmd.SubCommands(mainCmd)

	if err = mainCmd.Execute(); err != nil {
		logrus.Errorf("execute main cmd failed. error: %s", err)
		return
	}
}
