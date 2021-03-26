package lswagger

import (
	"bytes"
	"encoding/json"
	"github.com/go-openapi/spec"
	"github.com/haozzzzzzzz/go-tool/common/source"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type Swagger struct {
	spec.Swagger
}

func NewSwagger() (swg *Swagger) {
	swg = &Swagger{}
	swg.Paths = NewPaths()
	swg.Swagger.Swagger = "2.0"
	return
}

func LoadSwagger(jsonFile string) (swag *Swagger, err error) {
	swag = NewSwagger()
	err = swag.Load(jsonFile)
	if err != nil {
		logrus.Errorf("load swag failed. error: %s", err)
		return
	}
	return
}

func (m *Swagger) PathsAdd(key string, pathItem *spec.PathItem) {
	m.Paths.Paths[key] = *pathItem
}

func (m *Swagger) Load(jsonFile string) (err error) {
	content, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		logrus.Errorf("load swagger failed. error: %s", err)
		return
	}

	err = m.UnmarshalJSON(content)
	if err != nil {
		logrus.Errorf("load from json failed. error: %s", err)
		return
	}
	return
}

func (m *Swagger) Merge(otherSwagger *Swagger) {
	for uri, pathItem := range otherSwagger.Paths.Paths {
		m.PathsAdd(uri, &pathItem)
	}
}

func (m *Swagger) Output(pretty bool) (output []byte, err error) {
	output, err = m.Swagger.MarshalJSON()
	if nil != err {
		logrus.Errorf("swagger marshal json failed. error: %s.", err)
		return
	}

	if !pretty {
		return
	}

	var buf bytes.Buffer
	err = json.Indent(&buf, output, "", "\t")
	if nil != err {
		logrus.Errorf("json indent swagger json bytes failed. error: %s.", err)
		return
	}

	output = buf.Bytes()
	return
}

func (m *Swagger) SaveFile(fileName string, pretty bool) (err error) {
	out, err := m.Output(pretty)
	if nil != err {
		logrus.Errorf("get spec output failed. error: %s.", err)
		return
	}

	err = ioutil.WriteFile(fileName, out, source.ProjectFileMode)
	if nil != err {
		logrus.Errorf("save spec to file failed. error: %s.", err)
		return
	}

	return
}

func NewPaths() *spec.Paths {
	return &spec.Paths{
		Paths: make(map[string]spec.PathItem),
	}
}
