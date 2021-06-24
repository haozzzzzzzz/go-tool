package parser

import (
	"fmt"
	"github.com/haozzzzzzzz/go-rapid-development/v2/utils/file"
	"github.com/haozzzzzzzz/go-tool/common/source"

	"os"

	"github.com/sirupsen/logrus"
)

type ApiParser struct {
	ServiceDir string // 服务根目录
	ApiDir     string // 服务下api目录. api_dir和service_dir可以一样
}

func NewApiParser(
	serviceDir string,
	apiDir string, // 指定api目录
) (apiParser *ApiParser, err error) {
	if apiDir == "" { // 如果不指定，则默认为当前目录下的api目录
		apiDir = fmt.Sprintf("%s/api", serviceDir)
	}

	if !file.PathExists(apiDir) {
		err = os.MkdirAll(apiDir, source.ProjectDirMode)
		if nil != err {
			logrus.Errorf("mkdir %s failed. error: %s.", apiDir, err)
			return
		}
	}

	apiParser = &ApiParser{
		ServiceDir: serviceDir,
		ApiDir:     apiDir,
	}

	return
}

func NewApiParserSpecify(apiDir string) (apiParser *ApiParser, err error) {
	apiParser = &ApiParser{
		ServiceDir: apiDir,
		ApiDir:     apiDir,
	}
	return
}

func (m *ApiParser) apiUrlKey(uri string, method string) string {
	return fmt.Sprintf("%s_%s", uri, method)
}
