/*
This package parses api items to swagger specification
*/
package parser

import (
	"github.com/haozzzzzzzz/go-tool/common/source"
	"strings"

	"fmt"

	"github.com/go-openapi/spec"
	"github.com/haozzzzzzzz/go-rapid-development/v2/api/request"
	"github.com/haozzzzzzzz/go-tool/lib/lswagger"
	"github.com/sirupsen/logrus"
)

type SwaggerSpec struct {
	apis                   []*ApiItem
	Swagger                *lswagger.Swagger
	swaggerSchemaConverter *ITypeSwaggerSchemaConverter
}

func NewSwaggerSpec() (swgSpec *SwaggerSpec) {
	swgSpec = &SwaggerSpec{
		apis:                   make([]*ApiItem, 0),
		Swagger:                lswagger.NewSwagger(),
		swaggerSchemaConverter: NewITypeSwaggerSchemaConverter(),
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

	pathApis := make(map[string][]*ApiItem)
	// parse apis
	for _, api := range m.apis {
		paths := api.RelativePaths
		for _, path := range paths { // if api has handler with multi paths, gen spec for each path
			_, ok := pathApis[path]
			if !ok {
				pathApis[path] = make([]*ApiItem, 0)
			}

			pathApis[path] = append(pathApis[path], api)
		}
	}

	for path, apis := range pathApis {
		err = m.parsePathApis(path, apis)
		if nil != err {
			logrus.Errorf("parse path apis failed. error: %s.", err)
			return
		}
	}

	// 生成model definitions
	m.setDefinitions()
	return
}

func (m *SwaggerSpec) parsePathApis(path string, apis []*ApiItem) (err error) {
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
	pathItem := &spec.PathItem{}

	// https://swagger.io/specification/v2/#parameterObject
	// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/2.0.md#items-object

	for _, api := range apis {
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
				params := NotBodyFieldParameters("path", pathField, nil)
				for _, param := range params {
					param.Required = true
					operation.Parameters = append(operation.Parameters, *param)
				}
			}
		}

		// header data
		if api.HeaderData != nil {
			for _, headerField := range api.HeaderData.Fields {
				params := NotBodyFieldParameters("header", headerField, nil)
				for _, param := range params {
					operation.Parameters = append(operation.Parameters, *param)
				}

			}
		}

		// query data
		if api.QueryData != nil {
			for _, queryField := range api.QueryData.Fields {
				params := NotBodyFieldParameters("query", queryField, nil)
				for _, param := range params {
					operation.Parameters = append(operation.Parameters, *param)
				}

			}
		}

		// post data
		if api.PostData != nil {
			body := &spec.Parameter{}
			body.In = "body"
			body.Name = "body"
			postTypeName := api.PostData.TypeName()
			postTypeDesc := api.PostData.TypeDescription()
			body.Description = postTypeName
			if postTypeDesc != "" {
				body.Description += "\n" + postTypeDesc
			}

			body.Required = true
			body.Schema = m.swaggerSchemaConverter.ToSwaggerSchema(api.PostData, []source.IType{})

			operation.Parameters = append(operation.Parameters, *body)
		}

		// response data
		successResponse := spec.Response{}
		successResponse.Description = "success"
		if api.RespData != nil {
			// wrap data for ginbuilder.HandleFunc
			if api.ApiHandlerFuncType == ApiHandlerFuncTypeGinbuilderHandleFunc {
				successResponse.Schema = m.swaggerSchemaConverter.ToSwaggerSchema(SuccessResponseStructType(api.RespData), []source.IType{})
			} else {
				successResponse.Schema = m.swaggerSchemaConverter.ToSwaggerSchema(api.RespData, []source.IType{})
			}
		}

		operation.Responses = &spec.Responses{
			ResponsesProps: spec.ResponsesProps{
				StatusCodeResponses: map[int]spec.Response{
					200: successResponse,
				},
			},
		}

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

	}

	m.Swagger.PathsAdd(path, pathItem)
	return
}

// Apis set apis for building swagger spec
func (m *SwaggerSpec) Apis(apis []*ApiItem) {
	m.apis = apis
}

// Host set swagger host params
func (m *SwaggerSpec) Host(host string) {
	m.Swagger.Host = host
}

// Schemes set swagger schemes params
func (m *SwaggerSpec) Schemes(schemes []string) {
	m.Swagger.Schemes = schemes
}

// Info set swagger info params
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

// model definitions
func (m *SwaggerSpec) setDefinitions() {
	m.Swagger.Definitions = m.swaggerSchemaConverter.SwaggerDefinitions()
	return
}

// SaveToFile save swagger spec to file
func (m *SwaggerSpec) SaveToFile(fileName string, pretty bool) (err error) {
	return m.Swagger.SaveFile(fileName, pretty)
}

// NotBodyFieldParameters 非body里声明的类型参数
// https://swagger.io/specification/v2/#parameterObject
// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/2.0.md#items-object
func NotBodyFieldParameters(
	in string,
	field *source.Field,
	fieldTypeStack []source.IType, // 字段类型解析栈, 避免环形解析
) (parameters []*spec.Parameter) {
	if fieldTypeStack == nil {
		fieldTypeStack = make([]source.IType, 0)
	}

	// 检查是否有循环引用的类型
	hasITypeLoop := false
	for _, stackType := range fieldTypeStack {
		if stackType == field.TypeSpec {
			hasITypeLoop = true
			break
		}
	}

	if hasITypeLoop {
		logrus.Warnf("swagger pase params field exit because found ref loop . field : %+v, stack: %+v", field, fieldTypeStack)
		return
	}

	fieldTypeStack = append(fieldTypeStack, field.TypeSpec)

	parameters = make([]*spec.Parameter, 0)
	switch typeSpec := field.TypeSpec.(type) {
	case *source.StructType: // params embedded in struct
		for _, subField := range typeSpec.Fields {
			parameters = append(parameters, NotBodyFieldParameters(in, subField, fieldTypeStack)...)
		}

	default: // basic type or array
		param := BasicITypeToSwaggerParameter(in, field)
		parameters = append(parameters, param)

	}

	return
}

func BasicITypeToSwaggerParameter(in string, field *source.Field) (parameter *spec.Parameter) {
	items := BasicITypeToSwaggerNotBodyItemsObject(field.TypeSpec)
	parameter = &spec.Parameter{
		CommonValidations: items.CommonValidations,
		SimpleSchema:      items.SimpleSchema,
		VendorExtensible:  items.VendorExtensible,
		ParamProps: spec.ParamProps{
			Name:        field.TagJson(),
			In:          in,
			Description: field.Description,
			Required:    field.Required(),
		},
	}
	return
}

// BasicITypeToSwaggerNotBodyItemsObject array or basic type to spec parameter item
func BasicITypeToSwaggerNotBodyItemsObject(iType source.IType) (items *spec.Items) {
	switch fieldType := iType.(type) {
	case *source.ArrayType:
		items = spec.NewItems()
		items.Type = "array"
		items.Items = BasicITypeToSwaggerNotBodyItemsObject(fieldType.EltSpec)

	case *source.BasicType:
		items = spec.NewItems()
		items.Type = BasicTypeToSwaggerSchemaType(fieldType.TypeName())

	default:
		logrus.Warnf("not-in-body items type %+v has not supported yet", iType)

	}

	return
}

// BasicTypeToSwaggerSchemaType transform basic type to swagger schema type
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

type ITypeSwaggerSchemaConverter struct {
	structSchemaMap map[*source.StructType]*spec.Schema
	structDefIdMap  map[*source.StructType]string // struct_type -> swagger definition model id
}

func NewITypeSwaggerSchemaConverter() *ITypeSwaggerSchemaConverter {
	return &ITypeSwaggerSchemaConverter{
		structSchemaMap: map[*source.StructType]*spec.Schema{},
		structDefIdMap:  map[*source.StructType]string{},
	}
}

func (m *ITypeSwaggerSchemaConverter) ToSwaggerSchema(
	iType source.IType,
	fieldTypeStack []source.IType, // 字段类型解析栈, 避免环形解析
) (
	schema *spec.Schema,
) {
	if fieldTypeStack == nil {
		fieldTypeStack = make([]source.IType, 0)
	}

	switch t := iType.(type) {
	case *source.StructType:
		if t.Underlying != nil {
			iType = t.Underlying // 使用内部的类
		}
	}

	// 检查是否有循环引用的类型
	hasITypeLoop := false
	for _, stackType := range fieldTypeStack {
		if stackType == iType {
			hasITypeLoop = true
			break
		}
	}

	schema = &spec.Schema{}
	var subStack []source.IType
	if !hasITypeLoop {
		subStack = append(fieldTypeStack, iType)
	}

	switch iType.(type) {
	case *source.StructType:
		structType := iType.(*source.StructType)
		schema.Type = []string{"object"}
		if hasITypeLoop { // 结束循环
			return
		}

		schema = m.structTypeSwaggerSchema(structType, subStack)

	case *source.MapType:
		mapType := iType.(*source.MapType)
		schema.Type = []string{"object"}
		if hasITypeLoop {
			return
		}

		schema.AdditionalProperties = &spec.SchemaOrBool{}
		schema.AdditionalProperties.Schema = m.ToSwaggerSchema(mapType.ValueSpec, subStack)

	case *source.ArrayType:
		arrayType := iType.(*source.ArrayType)
		schema.Type = []string{"array"}
		if hasITypeLoop {
			return
		}

		schema.Items = &spec.SchemaOrArray{}
		schema.Items.Schema = m.ToSwaggerSchema(arrayType.EltSpec, subStack)

	case *source.InterfaceType:
		schema.Type = []string{"object"}

	case *source.BasicType:
		basicType := iType.(*source.BasicType)
		schema = NewBasicSwaggerSchema(basicType.Name)

	default:
		fmt.Println("unsupported itype for swagger schema")

	}

	return
}

func NewBasicSwaggerSchema(basicTypeName string) (schema *spec.Schema) {
	schema = &spec.Schema{}
	schema.Type = []string{
		BasicTypeToSwaggerSchemaType(basicTypeName),
	}
	return
}

// translate struct type to swagger's object definition schema
func (m *ITypeSwaggerSchemaConverter) structTypeSwaggerSchema(
	structType *source.StructType,
	subStack []source.IType,
) (refSchema *spec.Schema) {
	// 查询是否已经生成过model
	refSchema, ok := m.structSchemaMap[structType]
	if ok {
		return
	}

	schema := &spec.Schema{}
	schema.Type = []string{"object"}
	schema.Required = make([]string, 0) // object sub required field
	schema.Properties = make(map[string]spec.Schema)

	defer func() {
		refSchema = &spec.Schema{}

		defId := m.genDefId(structType)
		refLink := fmt.Sprintf("#/definitions/%s", defId)
		refSchema.Ref = spec.MustCreateRef(refLink)

		m.structSchemaMap[structType] = schema
		m.structDefIdMap[structType] = defId
	}()

	// 自身扩展字段
	for _, field := range structType.Fields {
		_, apiDocTags := field.Tags.ApiDoc()
		if apiDocTags.Skip {
			continue
		}

		if field.Embedded || !field.Exported {
			continue
		}

		// 非嵌入的、公开访问的field
		jsonName := field.TagJson()
		var fieldSchema *spec.Schema
		if apiDocTags.StrType != "" {
			fieldSchema = NewBasicSwaggerSchema(apiDocTags.StrType)
		} else {
			fieldSchema = m.ToSwaggerSchema(field.TypeSpec, subStack)
		}

		fieldSchema.Description = field.Description
		if field.Required() {
			// declare required field to object
			schema.Required = append(schema.Required, jsonName)
		}

		schema.Properties[jsonName] = *fieldSchema
	}

	// 内嵌对象字段
	for _, embeddedField := range structType.Embedded {
		_, apiDocTags := embeddedField.Tags.ApiDoc()
		if apiDocTags.Skip { // 跳过字段
			return
		}

		_, ok := embeddedField.TypeSpec.(*source.StructType)
		if !ok {
			continue
		}

		var embeddedSchema *spec.Schema
		if apiDocTags.StrType != "" {
			embeddedSchema = NewBasicSwaggerSchema(apiDocTags.StrType)
		} else {
			embeddedSchema = m.ToSwaggerSchema(embeddedField.TypeSpec, subStack)
		}

		if embeddedSchema == nil {
			logrus.Errorf("parse embedded struct type to swagger schema failed. embedded type: %s", embeddedField.TypeName)
			continue
		}

		// required
		schema.AddRequired(embeddedSchema.Required...)

		// merge field
		for fieldName, fieldSchema := range embeddedSchema.Properties {
			_, ok := schema.Properties[fieldName]
			if ok {
				//logrus.Warnf("conflict embedded struct field name. fieldName: %s, embedded struct: %s", fieldName, embeddedField.TypeName)
				continue
			}

			schema.Properties[fieldName] = fieldSchema
		}
	}

	return
}

func (m *ITypeSwaggerSchemaConverter) genDefId(structType *source.StructType) (defId string) {
	typeName := structType.Name
	strAddr := fmt.Sprintf("%p", structType)
	return fmt.Sprintf("%s-%s", typeName, strAddr)
}

func (m *ITypeSwaggerSchemaConverter) SwaggerDefinitions() (definitions spec.Definitions) {
	definitions = make(spec.Definitions)
	for structType, schema := range m.structSchemaMap {
		defId := m.structDefIdMap[structType]
		definitions[defId] = *schema
	}
	return
}
