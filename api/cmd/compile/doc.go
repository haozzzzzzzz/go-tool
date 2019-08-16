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
	var outputDir string

	var host string
	var version string
	var contactName string
	var notMod bool
	var serviceName string
	var serviceDescription string

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

			outputDir, err = filepath.Abs(outputDir)
			if nil != err {
				logrus.Errorf("get abs output path failed. error: %s.", err)
				return
			}

			// api parser
			apiParser := parser.NewApiParser(serviceDir)
			apis, err := apiParser.ScanApis(true, !notMod)
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
			err = swaggerSpec.SaveToFile(fmt.Sprintf("%s/swagger.json", outputDir))
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
	flags.StringVarP(&serviceName, "name", "n", "service name", "service name")
	flags.StringVarP(&serviceDescription, "desc", "d", "description", "description")
	flags.StringVarP(&host, "host", "H", "", "api host")
	flags.StringVarP(&version, "version", "v", "1.0", "api version")
	flags.StringVarP(&contactName, "contact_name", "c", "", "contact name")
	flags.BoolVarP(&notMod, "not_mod", "N", false, "not mod")
	flags.StringVarP(&outputDir, "output", "o", "./", "doc output path")

	return cmd
}

func GenerateCommentSwagger() (cmd *cobra.Command) {
	var dir string
	var outputDir string
	var host string
	cmd = &cobra.Command{
		Use:   "com_swagger",
		Short: "generate swagger specification from comment",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if dir == "" || outputDir == "" {
				logrus.Errorf("invalid params")
				return
			}

			dir, err = filepath.Abs(dir)
			if nil != err {
				logrus.Errorf("get abs dir failed. dir: %s, error: %s.", dir, err)
				return
			}

			outputDir, err = filepath.Abs(outputDir)
			if nil != err {
				logrus.Errorf("get abs output path failed. error: %s.", err)
				return
			}

			title, description, version, contact, apis, err := parser.ParseApisFromComments(dir)
			if nil != err {
				logrus.Errorf("parse failed. error: %s.", err)
				return
			}

			swaggerSpec := parser.NewSwaggerSpec()
			swaggerSpec.Info(
				title,
				description,
				version,
				contact,
			)
			swaggerSpec.Host(host)
			swaggerSpec.Apis(apis)
			swaggerSpec.Schemes([]string{"http", "https"})
			err = swaggerSpec.ParseApis()
			err = swaggerSpec.SaveToFile(fmt.Sprintf("%s/swagger.json", outputDir))
			if nil != err {
				logrus.Errorf("save swagger spec to file failed. error: %s.", err)
				return
			}

			return
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&dir, "dir", "d", "./", "search directory")
	flags.StringVarP(&outputDir, "output", "o", "./", "doc output dir")
	flags.StringVarP(&host, "host", "H", "", "host")
	return
}
