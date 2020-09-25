package cmd

import (
	"github.com/haozzzzzzzz/go-tool/ws/doc"
	"github.com/haozzzzzzzz/go-tool/ws/parse"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

func DocCmd() (command *cobra.Command) {
	var fileFormat string
	var outputFilePath string
	var rootDir string
	command = &cobra.Command{
		Use: "doc",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			fileFormat = strings.ToLower(fileFormat)
			if fileFormat == "" {
				fileFormat = "json"
				return
			}

			rootDir, err = filepath.Abs(rootDir)
			if err != nil {
				logrus.Errorf("get root rootDir abs path failed. error: %s", err)
				return
			}

			absOutputFilePath, err := filepath.Abs(outputFilePath)
			if err != nil {
				logrus.Errorf("get absolute file path failed. filepath: %s, error: %s", outputFilePath, err)
				return
			}

			parser := parse.NewWsTypesParser(rootDir)
			err = parser.ParseWsTypes()
			if err != nil {
				logrus.Errorf("parser ws types failed. error: %s", err)
				return
			}

			err = doc.WriteDoc(parser.WsTypes().Output(), fileFormat, absOutputFilePath)
			if err != nil {
				logrus.Errorf("write ws doc failed. file_format: %s, file_path: %s, error: %s", fileFormat, absOutputFilePath, err)
				return
			}

			logrus.Infoln("finish")
		},
	}

	flags := command.Flags()
	flags.StringVarP(&rootDir, "root_dir", "d", "./", "source root dir")
	flags.StringVarP(&fileFormat, "format", "f", "json", "doc file format")
	flags.StringVarP(&outputFilePath, "output", "o", "./ws_doc.json", "doc output file path")
	return
}
