package precompile

import (
	"fmt"
	yaml2 "github.com/haozzzzzzzz/go-rapid-development/utils/yaml"
	"github.com/haozzzzzzzz/go-tool/code/com/precompiler"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"strings"
)

func CommandPrecompile() (cmd *cobra.Command) {
	var path string
	var paramsPath string
	cmd = &cobra.Command{
		Use:   "precompile",
		Short: "precompile --path filepath [ --params_file params_filepath ] [ key1=val1 key2=val2 ]",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if path == "" {
				logrus.Errorf("require path")
				return
			}

			path, err = filepath.Abs(path)
			if nil != err {
				logrus.Errorf("get abs filepath failed. path: %s, error: %s.", path, err)
				return
			}

			params := make(map[string]interface{})

			if paramsPath != "" {
				paramsPath, err = filepath.Abs(paramsPath)
				if nil != err {
					logrus.Errorf("get abs params filepath failed. error: %s.", err)
					return
				}

				err = yaml2.ReadYamlFromFile(paramsPath, &params)
				if nil != err {
					logrus.Errorf("read yaml from params file failed. error: %s.", err)
					return
				}
			}

			if len(args) > 0 {
				strYamlText := ""
				for _, strPair := range args {
					pair := strings.Split(strPair, "=")
					if len(pair) != 2 {
						continue
					}

					strYamlText += fmt.Sprintf("%s : %s \n", pair[0], pair[1])
				}

				err = yaml.Unmarshal([]byte(strYamlText), &params)
				if nil != err {
					logrus.Errorf("yaml unmarshal params failed. error: %s.", err)
					return
				}

			}

			if len(params) > 0 {
				logrus.Printf("params: %#v\n", params)
			}

			err = precompiler.Precompile(path, params)
			if nil != err {
				logrus.Errorf("precompile failed. error: %s.", err)
				return
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&path, "path", "p", "./", "precompile path")
	flags.StringVar(&paramsPath, "params", "", "params file path")
	return
}
