package compile

import (
	"os"
	"path/filepath"

	"github.com/haozzzzzzzz/go-tool/api/com/parser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// compile RESTful http api
func CommandApiCompile() *cobra.Command {
	var serviceDir string
	var apiDir string // 指定解析的接口目录. 如果不指定，则默认为{serviceDir}/api

	var notMod bool
	var cmd = &cobra.Command{
		Use:   "compile",
		Short: "api service compilation",
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			defer func() {
				if err != nil {
					os.Exit(1)
				}
			}()

			if serviceDir == "" {
				logrus.Errorf("service dir required")
				return
			}

			serviceDir, err = filepath.Abs(serviceDir)
			if nil != err {
				logrus.Errorf("get absolute service path failed. %s.", err)
				return
			}

			apiDir, err := filepath.Abs(apiDir)
			if err != nil {
				logrus.Errorf("get absolute api dir path failed. error: %s", err)
				return
			}

			// api parser
			apiParser, err := parser.NewApiParser(serviceDir, apiDir)
			if nil != err {
				logrus.Errorf("new api parser failed. error: %s.", err)
				return
			}

			_, apis, err := apiParser.ScanApis(false, false, !notMod)
			if nil != err {
				logrus.Errorf("Scan api failed. %s.", err)
				return
			}

			err = apiParser.GenerateRoutersSourceFile(apis)
			if nil != err {
				logrus.Errorf("generate routers source file failed. %s.", err)
				return
			}

		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&serviceDir, "path", "p", "./", "service path")
	flags.StringVarP(&apiDir, "api_dir", "a", "./api", "api dir")
	flags.BoolVarP(&notMod, "not_mod", "n", false, "not mod")

	return cmd
}
