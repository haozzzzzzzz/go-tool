package parser

import (
	"fmt"
	"github.com/haozzzzzzzz/go-rapid-development/v2/api/request"
	"github.com/haozzzzzzzz/go-rapid-development/v2/utils/uerrors"
	"github.com/haozzzzzzzz/go-tool/common/source"
	"github.com/haozzzzzzzz/go-tool/lib/gofmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

/*
Generate api router files
*/

func (m *ApiParser) GenerateRoutersSourceFile(apis []*ApiItem) (err error) {
	logrus.Info("Map apis ...")
	defer func() {
		if err == nil {
			logrus.Info("Map apis completed")
		}
	}()

	importsMap := make(map[string]string) // package_exported -> alias
	strRouters := make([]string, 0)

	for _, apiItem := range apis {
		// imports
		if apiItem.ApiFile.PackageRelAlias != "" { // 本目录下的不用加入
			if apiItem.ApiFile.PackageExportedPath == "" {
				err = uerrors.Newf("alias require exported path. alias: %s", apiItem.ApiFile.PackageRelAlias)
				return
			}

			importsMap[apiItem.ApiFile.PackageExportedPath] = apiItem.ApiFile.PackageRelAlias
		}

		// handle func binding
		strHandleFunc := apiItem.ApiHandlerFunc
		if apiItem.ApiFile.PackageRelAlias != "" {
			strHandleFunc = fmt.Sprintf("%s.%s", apiItem.ApiFile.PackageRelAlias, apiItem.ApiHandlerFunc)
		}

		handleFuncName := ""
		switch apiItem.ApiHandlerFuncType {
		case ApiHandlerFuncTypeGinHandlerFunc:
			handleFuncName = strHandleFunc
		case ApiHandlerFuncTypeGinbuilderHandleFunc:
			handleFuncName = fmt.Sprintf("%s.GinHandler", strHandleFunc)
		default:
			continue
		}

		for _, uri := range apiItem.RelativePaths {
			var str string
			switch apiItem.HttpMethod {
			case request.METHOD_ANY:
				str = fmt.Sprintf("    engine.Any(\"%s\", %s)", uri, handleFuncName)
			default:
				str = fmt.Sprintf("    engine.Handle(\"%s\", \"%s\", %s)", apiItem.HttpMethod, uri, handleFuncName)
			}

			if str == "" {
				logrus.Errorf("gen code for api item failed. api: %#v, error: %s.", apiItem, err)
				return
			}

			strRouters = append(strRouters, str)
		}

	}

	strImports := make([]string, 0)
	for expPath, alias := range importsMap {
		var str string
		if alias == "" {
			str = fmt.Sprintf("    %q", expPath)
		} else {
			str = fmt.Sprintf("    %s %q", alias, expPath)
		}
		strImports = append(strImports, str)
	}

	routersFileName := fmt.Sprintf("%s/routers.go", m.ApiDir)
	newRoutersText := fmt.Sprintf(routersFileText, strings.Join(strImports, "\n"), strings.Join(strRouters, "\n"))
	newRoutersText, err = gofmt.StrGoFmt(newRoutersText)
	if nil != err {
		logrus.Errorf("go fmt source failed. text: %s, error: %s.", newRoutersText, err)
		return
	}

	err = ioutil.WriteFile(routersFileName, []byte(newRoutersText), 0644)
	if nil != err {
		logrus.Errorf("write new routers file failed. \n%s.", err)
		return
	}

	return
}

func (m *ApiParser) SaveApisToFile(
	apis []*ApiItem,
	filePath string,
) (err error) {
	logrus.Info("Save apis ...")
	defer func() {
		if err == nil {
			logrus.Info("Save apis completed")
		}
	}()

	// save api.yaml
	byteYamlApis, err := yaml.Marshal(apis)
	if nil != err {
		logrus.Errorf("yaml marshal apis failed. %s.", byteYamlApis)
		return
	}

	err = ioutil.WriteFile(filePath, byteYamlApis, source.ProjectFileMode)
	if nil != err {
		logrus.Errorf("write apis.yaml failed. %s.", err)
		return
	}

	return
}

var routersFileText = `package api
import (
"github.com/gin-gonic/gin"
%s
)

// BindRouters 注意：BindRouters函数体内不能自定义添加任何声明，由api compile命令生成api绑定声明
func BindRouters(engine *gin.Engine) (err error) {
%s
	return
}
`
