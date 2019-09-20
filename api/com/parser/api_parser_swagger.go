/**
This package parses api items to swagger specification
*/
package parser

import (
	"strings"

	"fmt"

	"io/ioutil"

	"bytes"
	"encoding/json"

	"github.com/go-openapi/spec"
	"github.com/haozzzzzzzz/go-rapid-development/api/request"
	"github.com/haozzzzzzzz/go-tool/api/com/project"
	"github.com/haozzzzzzzz/go-tool/lib/lswagger"
	"github.com/sirupsen/logrus"
)

type SwaggerSpec struct {
	commonParamsMap map[string]*ApiItemParams // dir -> common params
	apis            []*ApiItem
	Swagger         *lswagger.Swagger
}

func NewSwaggerSpec() (swgSpec *SwaggerSpec) {
	swgSpec = &SwaggerSpec{
		apis:    make([]*ApiItem, 0),
		Swagger: lswagger.NewSwagger(),
	}
	return
}

func (m *SwaggerSpec) ParseApis() (
	err error,
) {
	logrus.Info("Save swagger spec ...")
	defer func() {
		if nil != err {
			logrus.Errorf("Save swagger spec failed. %s", err)
			return
		} else {
			logrus.Info("Save swagger spec finish")
		}
	}()

	// parse apis
	for _, api := range m.apis {
		paths := api.RelativePaths
		for _, path := range paths { // if api has handler with multi paths, gen spec for each path
			err = m.parseApi(path, api)
			if nil != err {
				logrus.Errorf("swagger spec parse api failed. error: %s.", err)
				return
			}

			err = m.MergeFromCommonParams(api)
			if nil != err {
				logrus.Errorf("merge from common params failed. error: %s.", err)
				return
			}

		}
	}

	return
}

func (m *SwaggerSpec) MergeFromCommonParams(apiItem *ApiItem) (err error) {
	if m.commonParamsMap == nil || apiItem == nil {
		return
	}

	pkgDir := apiItem.ApiFile.PackageDir
	matchedCommonParams := make([]*ApiItemParams, 0)
	for dir, commonParams := range m.commonParamsMap {
		if strings.Contains(pkgDir, dir) {
			matchedCommonParams = append(matchedCommonParams, commonParams)
		}
	}

	if len(matchedCommonParams) == 0 {
		return
	}

	for _, commonParams := range matchedCommonParams {
		err = apiItem.ApiItemParams.MergeApiItemParams(commonParams)
		if nil != err {
			logrus.Errorf("api merge common params failed. %s", err)
			return
		}
	}

	return
}

func (m *SwaggerSpec) parseApi(path string, api *ApiItem) (err error) {
	// transform gin-style url path params "/:param" to swagger-style url param "{param}"
	subPaths := strings.Split(path, "/")
	for i, subPath := range subPaths {
		if strings.HasPrefix(subPath, ":") {
			subPath = strings.Replace(subPath, ":", "", 1)
			subPath = fmt.Sprintf("{%s}", subPath)
			subPaths[i] = subPath
		}
	}

	path = strings.Join(subPaths, "/")

	operation := &spec.Operation{}
	operation.ID = fmt.Sprintf("%s-%s", api.HttpMethod, path)
	operation.Consumes = []string{request.MIME_JSON}
	operation.Produces = []string{request.MIME_JSON}
	operation.Summary = api.Summary
	operation.Description = api.Description
	operation.Parameters = make([]spec.Parameter, 0)
	operation.Tags = api.Tags
	operation.Deprecated = api.Deprecated

	// uri data
	if api.UriData != nil {
		for _, pathField := range api.UriData.Fields {
			basicParameter := *FieldBasicParameter("path", pathField)
			basicParameter.Required = true // require uri data
			operation.Parameters = append(operation.Parameters, basicParameter)
		}
	}

	// header data
	if api.HeaderData != nil {
		for _, headerField := range api.HeaderData.Fields {
			operation.Parameters = append(operation.Parameters, *FieldBasicParameter("header", headerField))
		}
	}

	// query data
	if api.QueryData != nil {
		for _, queryField := range api.QueryData.Fields {
			operation.Parameters = append(operation.Parameters, *FieldBasicParameter("query", queryField))
		}
	}

	// post data
	if api.PostData != nil {
		body := &spec.Parameter{}
		body.In = "body"
		body.Name = api.PostData.Name
		body.Description = api.PostData.Description
		body.Required = true
		body.Schema = ITypeToSwaggerSchema(api.PostData)
		operation.Parameters = append(operation.Parameters, *body)
	}

	// response data
	successResponse := spec.Response{}
	successResponse.Description = "success"
	if api.RespData != nil {

		// wrap data for ginbuilder.HandleFunc
		if api.ApiHandlerFuncType == ApiHandlerFuncTypeGinbuilderHandleFunc {
			successResponse.Schema = ITypeToSwaggerSchema(SuccessResponseStructType(api.RespData))
		} else {
			successResponse.Schema = ITypeToSwaggerSchema(api.RespData)
		}

	}

	operation.Responses = &spec.Responses{
		ResponsesProps: spec.ResponsesProps{
			StatusCodeResponses: map[int]spec.Response{
				200: successResponse,
			},
		},
	}

	pathItem := &spec.PathItem{}

	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#pathsObject
	switch api.HttpMethod {
	case request.METHOD_GET:
		pathItem.Get = operation
	case request.METHOD_POST:
		pathItem.Post = operation
	case request.METHOD_PUT:
		pathItem.Put = operation
	case request.METHOD_DELETE:
		pathItem.Delete = operation
	case request.METHOD_OPTIONS:
		pathItem.Options = operation
	case request.METHOD_HEAD:
		pathItem.Head = operation
	case request.METHOD_PATCH:
		pathItem.Patch = operation

	case request.METHOD_ANY:
		operationGet := *operation
		operationGet.ID = fmt.Sprintf("%s-%s", request.METHOD_GET, path)

		operationPost := *operation
		operationPost.ID = fmt.Sprintf("%s-%s", request.METHOD_POST, path)

		operationPut := *operation
		operationPut.ID = fmt.Sprintf("%s-%s", request.METHOD_PUT, path)

		operationDelete := *operation
		operationDelete.ID = fmt.Sprintf("%s-%s", request.METHOD_DELETE, path)

		operationOptions := *operation
		operationOptions.ID = fmt.Sprintf("%s-%s", request.METHOD_OPTIONS, path)

		operationHead := *operation
		operationHead.ID = fmt.Sprintf("%s-%s", request.METHOD_HEAD, path)

		operationPatch := *operation
		operationPatch.ID = fmt.Sprintf("%s-%s", request.METHOD_PATCH, path)

		pathItem.Get = &operationGet
		pathItem.Post = &operationPost
		pathItem.Put = &operationPut
		pathItem.Delete = &operationDelete
		pathItem.Options = &operationOptions
		pathItem.Head = &operationHead
		pathItem.Patch = &operationPatch

	default:
		logrus.Warnf("not supported method for swagger spec. method: %s", api.HttpMethod)
	}

	m.Swagger.PathsAdd(path, pathItem)

	return
}

// set apis for building swagger spec
func (m *SwaggerSpec) Apis(commonParamsMap map[string]*ApiItemParams, apis []*ApiItem) {
	m.commonParamsMap = commonParamsMap
	m.apis = apis
}

// set swagger host params
func (m *SwaggerSpec) Host(host string) {
	m.Swagger.Host = host
}

// set swagger schemes params
func (m *SwaggerSpec) Schemes(schemes []string) {
	m.Swagger.Schemes = schemes
}

// set swagger info params
func (m *SwaggerSpec) Info(
	title string,
	description string,
	version string,
	contactName string,
) {
	m.Swagger.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       title,
			Description: description,
			Version:     version,
			Contact: &spec.ContactInfo{
				Name: contactName,
			},
		},
	}

	return
}

// save swagger spec to file
func (m *SwaggerSpec) SaveToFile(fileName string) (err error) {
	out, err := m.Output()
	if nil != err {
		logrus.Errorf("get spec output failed. error: %s.", err)
		return
	}

	err = ioutil.WriteFile(fileName, out, project.ProjectFileMode)
	if nil != err {
		logrus.Errorf("save spec to file failed. error: %s.", err)
		return
	}

	return
}

// output swagger spec bytes
func (m *SwaggerSpec) Output() (output []byte, err error) {
	output, err = m.Swagger.MarshalJSON()
	if nil != err {
		logrus.Errorf("swagger marshal json failed. error: %s.", err)
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

// query、path基础类型参数
func FieldBasicParameter(in string, field *Field) (parameter *spec.Parameter) {
	parameter = &spec.Parameter{}
	parameter.Name = field.TagJson()
	parameter.In = in
	parameter.Description = field.Description
	parameter.Required = field.Required()
	switch field.TypeSpec.(type) {
	case *BasicType:
		parameter.Type = BasicTypeToSwaggerSchemaType(field.TypeName)

	default:
		parameter.Type = BasicTypeToSwaggerSchemaType(field.TypeName)

	}

	return
}

// transform basic type to swagger schema type
func BasicTypeToSwaggerSchemaType(fieldType string) (swagType string) {
	switch fieldType {
	case "string":
		swagType = "string"

	case "bool":
		swagType = "boolean"

	default:
		if strings.Contains(fieldType, "float") {
			swagType = "number"
		} else {
			swagType = "integer"
		}
	}
	return
}

func ITypeToSwaggerSchema(iType IType) (schema *spec.Schema) {
	schema = &spec.Schema{}
	switch iType.(type) {
	case *StructType:
		structType := iType.(*StructType)
		schema.Type = []string{"object"}
		schema.Required = make([]string, 0)
		schema.Properties = make(map[string]spec.Schema)

		for _, field := range structType.Fields {
			jsonName := field.TagJson()
			fieldSchema := ITypeToSwaggerSchema(field.TypeSpec)
			fieldSchema.Description = field.Description
			if field.Required() {
				fieldSchema.Required = []string{jsonName}
			}

			schema.Properties[jsonName] = *fieldSchema
		}

	case *MapType:
		mapType := iType.(*MapType)
		schema.Type = []string{"object"}
		schema.AdditionalProperties = &spec.SchemaOrBool{}

		schema.AdditionalProperties.Schema = ITypeToSwaggerSchema(mapType.ValueSpec)

	case *ArrayType:
		arrayType := iType.(*ArrayType)
		schema.Type = []string{"array"}
		schema.Items = &spec.SchemaOrArray{}
		schema.Items.Schema = ITypeToSwaggerSchema(arrayType.EltSpec)

	case *InterfaceType:
		//interType := iType.(*InterfaceType)
		schema.Type = []string{"object"}

	case *BasicType:
		basicType := iType.(*BasicType)
		schemaType := BasicTypeToSwaggerSchemaType(basicType.Name)
		schema.Type = []string{schemaType}

	default:
		fmt.Println("unsported itype for swagger schema")
	}

	return
}
