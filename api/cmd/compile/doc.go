package compile

import (
	"fmt"
	"path/filepath"

	"github.com/haozzzzzzzz/go-tool/api/com/parser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func GenerateSwaggerSpecification() *cobra.Command {
	var serviceDir string // 服务目录
	var apiDir string     // 指定解析的接口目录. 如果不指定，则默认为{serviceDir}/api
	var outputDir string

	var host string
	var version string
	var contactName string
	var notMod bool
	var serviceName string
	var serviceDescription string
	var pretty bool

	var cmd = &cobra.Command{
		Use:   "swagger",
		Short: "swagger specification generate",
		Run: func(cmd *cobra.Command, args []string) {
			if serviceDir == "" {
				logrus.Errorf("service dir required")
				return
			}

			var err error
			serviceDir, err = filepath.Abs(serviceDir)
			if nil != err {
				logrus.Errorf("get absolute service path failed. err: %s.", err)
				return
			}

			apiDir, err := filepath.Abs(apiDir)
			if err != nil {
				logrus.Errorf("get absolute api dir path failed. error: %s", err)
				return
			}

			outputDir, err = filepath.Abs(outputDir)
			if nil != err {
				logrus.Errorf("get abs output path failed. error: %s.", err)
				return
			}

			// api parser
			apiParser, err := parser.NewApiParser(serviceDir, apiDir)
			if nil != err {
				logrus.Errorf("new api parser failed. error: %s.", err)
				return
			}

			_, apis, err := apiParser.ScanApis(true, true, !notMod)
			if nil != err {
				logrus.Errorf("Scan api failed. err: %s.", err)
				return
			}

			swaggerSpec := parser.NewSwaggerSpec()
			swaggerSpec.Info(
				serviceName,
				serviceDescription,
				version,
				contactName,
			)
			swaggerSpec.Host(host)
			swaggerSpec.Apis(apis)
			swaggerSpec.Schemes([]string{"http", "https"})
			err = swaggerSpec.ParseApis()
			err = swaggerSpec.SaveToFile(fmt.Sprintf("%s/swagger.json", outputDir), pretty)
			if nil != err {
				logrus.Errorf("save swagger spec to file failed. error: %s.", err)
				return
			}

			err = apiParser.SaveApisToFile(apis, fmt.Sprintf("%s/apis.yaml", outputDir))
			if nil != err {
				logrus.Errorf("save apis to file failed. err: %s.", err)
				return
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&serviceDir, "path", "p", "./", "service path")
	flags.StringVarP(&apiDir, "api_dir", "a", "./api", "api dir")
	flags.StringVarP(&serviceName, "name", "n", "service name", "service name")
	flags.StringVarP(&serviceDescription, "desc", "d", "description", "description")
	flags.StringVarP(&host, "host", "H", "", "api host")
	flags.StringVarP(&version, "version", "v", "1.0", "api version")
	flags.StringVarP(&contactName, "contact_name", "c", "", "contact name")
	flags.BoolVarP(&notMod, "not_mod", "N", false, "not mod")
	flags.StringVarP(&outputDir, "output", "o", "./", "doc output path")
	flags.BoolVarP(&pretty, "pretty", "P", false, "pretty swagger.json output")

	return cmd
}
