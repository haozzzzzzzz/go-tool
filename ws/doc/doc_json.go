package doc

import (
	"github.com/haozzzzzzzz/go-rapid-development/utils/ujson"
	"github.com/haozzzzzzzz/go-tool/api/com/project"
	"github.com/haozzzzzzzz/go-tool/ws/parse"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func init() {
	BindFileFormatWriter([]string{"json"}, WriteDocJson)
}

func WriteDocJson(
	wsTypes *parse.WsTypes,
	format string,
	filepath string,
) (err error) {
	bDoc, err := ujson.MarshalPretty(wsTypes)
	if err != nil {
		logrus.Errorf("marshal pretty failed. error: %s", err)
		return
	}

	err = ioutil.WriteFile(filepath, bDoc, project.ProjectFileMode)
	if err != nil {
		logrus.Errorf("write json doc file failed. error: %s", err)
		return
	}

	return
}
