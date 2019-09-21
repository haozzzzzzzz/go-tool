package parser

import (
	"fmt"
	"github.com/haozzzzzzzz/go-rapid-development/utils/uerrors"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

type Field struct {
	Name        string            `json:"name" yaml:"name"`
	TypeName    string            `json:"type_name" yaml:"type_name"`
	Tags        map[string]string `json:"tags" yaml:"tags"`
	TypeSpec    IType             `json:"type_spec" form:"type_spec"`
	Description string            `json:"description" form:"description"`
}

func (m *Field) TagJsonOrName() (name string) {
	name = m.TagJson()
	if name == "" {
		name = m.Name
	}
	return
}

func (m *Field) TagJson() (name string) {
	return m.Tags["json"]
}

func (m *Field) TagForm() (name string) {
	return m.Tags["json"]
}

func (m *Field) Required() (required bool) {
	strBind := m.Tags["binding"]
	if strings.Contains(strBind, "required") {
		required = true
	}
	return
}

func NewField() *Field {
	return &Field{
		Tags: make(map[string]string),
	}
}

// type interface
type IType interface {
	TypeName() string
	TypeDescription() string
}

// 类型分类
const TypeClassBasicType = "basic"
const TypeClassStructType = "struct"
const TypeClassMapType = "map"
const TypeClassArrayType = "array"
const TypeClassInterfaceType = "interface"

// 标准类型
type BasicType struct {
	TypeClass   string `json:"type_class" yaml:"type_class"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

func (m BasicType) TypeName() string {
	return m.Name
}

func (m BasicType) TypeDescription() string {
	return m.Description
}

func NewBasicType(name string) *BasicType {
	return &BasicType{
		Name:      name,
		TypeClass: TypeClassBasicType,
	}
}

// struct
type StructType struct {
	TypeClass   string            `json:"type_class" yaml:"type_class"`
	Name        string            `json:"name" yaml:"name"`
	Fields      []*Field          `json:"fields" yaml:"fields"`
	mField      map[string]*Field `json:"-" yaml:"-"`
	Description string            `json:"description" yaml:"description"`
}

func (m *StructType) TypeName() string {
	return m.Name
}

func (m *StructType) TypeDescription() string {
	return m.Description
}

func (m *StructType) AddFields(fields ...*Field) (err error) {
	for _, field := range fields {
		if field.Name == "" {
			err = uerrors.Newf("struct field require name")
			return
		}

		fieldName := field.TagJsonOrName()
		_, ok := m.mField[fieldName]
		if ok {
			continue // skip same
		}

		m.Fields = append(m.Fields, field)
		m.mField[fieldName] = field
	}

	return
}

func NewStructType() *StructType {
	return &StructType{
		TypeClass: TypeClassStructType,
		Name:      TypeClassStructType,
		Fields:    make([]*Field, 0),
		mField:    make(map[string]*Field),
	}
}

// map
type MapType struct {
	TypeClass   string `json:"type_class" yaml:"type_class"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	KeySpec     IType  `json:"key" yaml:"key"`
	ValueSpec   IType  `json:"value_spec" yaml:"value_spec"`
}

func (m *MapType) TypeName() string {
	return m.Name
}

func (m *MapType) TypeDescription() string {
	return m.Description
}

func NewMapType() *MapType {
	return &MapType{
		TypeClass: TypeClassMapType,
		Name:      TypeClassMapType,
	}
}

// array
type ArrayType struct {
	TypeClass   string `json:"type_class" yaml:"type_class"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Len         int64  `json:"len" yaml:"len"`
	EltName     string `json:"elt_name" yaml:"elt_name"`
	EltSpec     IType  `json:"elt_spec" yaml:"elt_spec"`
}

func (m *ArrayType) TypeName() string {
	return m.Name
}

func (m *ArrayType) TypeDescription() string {
	return m.Description
}

func NewArrayType() *ArrayType {
	return &ArrayType{
		TypeClass: TypeClassArrayType,
		Name:      TypeClassArrayType,
	}
}

// interface
type InterfaceType struct {
	TypeClass string `json:"type_class" yaml:"type_class"`
}

func NewInterfaceType() *InterfaceType {
	return &InterfaceType{
		TypeClass: TypeClassInterfaceType,
	}
}

func (m *InterfaceType) TypeName() string {
	return m.TypeClass
}

func (m *InterfaceType) TypeDescription() string {
	return m.TypeClass
}

type ApiItemParams struct {
	HeaderData *StructType `json:"header_data" yaml:"header_data"`
	UriData    *StructType `json:"uri_data" yaml:"uri_data"`
	QueryData  *StructType `json:"query_data" yaml:"query_data"`
	PostData   IType       `json:"post_data" yaml:"post_data"`
	RespData   IType       `json:"response_data" yaml:"response_data"`
}

func NewApiItemParams() *ApiItemParams {
	return &ApiItemParams{
		HeaderData: NewStructType(),
		UriData:    NewStructType(),
		QueryData:  NewStructType(),
		PostData:   NewStructType(), // default
		RespData:   NewStructType(), // default
	}
}

func (m *ApiItemParams) MergeApiItemParams(items ...*ApiItemParams) (err error) {
	for _, item := range items {
		if item == nil {
			continue
		}

		if item.HeaderData != nil && len(item.HeaderData.Fields) > 0 {
			if m.HeaderData == nil {
				m.HeaderData = NewStructType()
			}

			err = m.HeaderData.AddFields(item.HeaderData.Fields...)
			if nil != err {
				logrus.Errorf("add header data fields failed. error: %s.", err)
				return
			}
		}

		if item.UriData != nil && len(item.UriData.Fields) > 0 {
			if m.UriData == nil {
				m.UriData = NewStructType()
			}

			err = m.UriData.AddFields(item.UriData.Fields...)
			if nil != err {
				logrus.Errorf("add uri data fields failed. error: %s.", err)
				return
			}
		}

		if item.QueryData != nil && len(item.QueryData.Fields) > 0 {
			if m.QueryData == nil {
				m.QueryData = NewStructType()
			}

			err = m.QueryData.AddFields(item.QueryData.Fields...)
			if nil != err {
				logrus.Errorf("add query data fields failed. error: %s.", err)
				return
			}
		}

		if item.PostData != nil {
			if m.PostData == nil {
				m.PostData = NewStructType()
			}

			switch item.PostData.(type) {
			case *StructType:
				itemPostStruct := item.RespData.(*StructType)
				if len(itemPostStruct.Fields) == 0 {
					break
				}

				mPostStruct, ok := m.PostData.(*StructType)
				if !ok {
					logrus.Warnf("can not merge post data, because of different type. %s %s", reflect.TypeOf(item.PostData), reflect.TypeOf(m.PostData))
					break
				}

				err = mPostStruct.AddFields(itemPostStruct.Fields...)
				if nil != err {
					logrus.Errorf("add post data fields failed. error: %s.", err)
					return
				}

			}
		}

		if item.RespData != nil {
			if m.RespData == nil {
				m.RespData = NewStructType()
			}

			switch item.RespData.(type) {
			case *StructType: // only merge struct type
				itemRespStruct := item.RespData.(*StructType)
				if itemRespStruct == nil {
					break
				}

				mRespStruct, ok := m.RespData.(*StructType)
				if ok {
					err = mRespStruct.AddFields(itemRespStruct.Fields...)
					if nil != err {
						logrus.Errorf("add resp data fields failed. error: %s.", err)
						return
					}

				} else {
					logrus.Warnf("can not merge resp data, because of different type. %s %s", reflect.TypeOf(item.RespData), reflect.TypeOf(m.RespData))

				}

			default:
				logrus.Warnf("only struct type resp data can be merged. type: %s", reflect.TypeOf(item.RespData))

			}

		}

	}
	return
}

func MergeApiItemParams(items ...*ApiItemParams) (params *ApiItemParams, err error) {
	params = NewApiItemParams()
	err = params.MergeApiItemParams(items...)
	if nil != err {
		logrus.Errorf("merge api items params failed. error: %s.", err)
		return
	}
	return
}

const ApiHandlerFuncTypeGinbuilderHandleFunc uint8 = 1
const ApiHandlerFuncTypeGinHandlerFunc uint8 = 2

type ApiItem struct {
	ApiItemParams

	ApiHandlerFuncType uint8  `json:"api_handler_func_type" yaml:"api_handler_func_type"`
	ApiHandlerFunc     string `validate:"required" json:"api_handler_func" yaml:"api_handler_func"` // handler 函数名

	ApiFile *ApiFile `json:"api_file" yaml:"api_file"` // api文件

	HttpMethod    string   `validate:"required" json:"http_method" yaml:"http_method"`
	RelativePaths []string `validate:"required" json:"relative_paths" yaml:"relative_paths"`

	// swagger
	Summary     string   `json:"summary" yaml:"summary"`
	Description string   `json:"description" yaml:"description"`
	Tags        []string `json:"tags" yaml:"tags"`             // 标签
	Deprecated  bool     `json:"deprecated" yaml:"deprecated"` // 是否是弃用的
}

func (m *ApiItem) PackageFuncName() string {
	return fmt.Sprintf("%s.%s", m.ApiFile.PackageRelAlias, m.ApiHandlerFunc)
}

func (m *ApiItem) MergeInfoFromApiInfo() {
	if m.ApiFile == nil {
		return
	}

	if len(m.ApiFile.Tags) > 0 {
		m.Tags = append(m.Tags, m.ApiFile.Tags...)
	}

	if m.ApiFile.Deprecated == true {
		m.Deprecated = true
	}

	return
}

func (m *ApiItem) MergeInfoFomCommentTags(tags *CommentTags) {
	if m.Summary == "" {
		m.Summary = tags.Summary
	}

	if m.Description == "" {
		m.Description = tags.Description
	}

	if m.HttpMethod == "" {
		m.HttpMethod = tags.LineTagDocHttpMethod
	}

	if len(tags.LineTagDocRelativePaths) > 0 {
		if m.RelativePaths == nil {
			m.RelativePaths = make([]string, 0)
		}
		m.RelativePaths = append(m.RelativePaths, tags.LineTagDocRelativePaths...)
	}

	if len(tags.LineTagDocTags) > 0 {
		if m.Tags == nil {
			m.Tags = make([]string, 0)
		}
		m.Tags = append(m.Tags, tags.LineTagDocTags...)
	}

	if tags.Deprecated {
		m.Deprecated = true
	}
}

// 成功返回结构
//type Response struct {
//	ReturnCode uint32      `json:"ret"`
//	Message    string      `json:"msg"`
//	Data       interface{} `json:"data"`
//}
func SuccessResponseStructType(
	respData IType,
) (successResp *StructType) {
	successResp = NewStructType()
	successResp.Name = "SuccessResponse"
	successResp.Description = "api success response"
	successResp.Fields = []*Field{
		{
			Name:        "ReturnCode",
			TypeName:    "uint32",
			Description: "result code",
			Tags: map[string]string{
				"json": "ret",
			},
			TypeSpec: NewBasicType("uint32"),
		},
		{
			Name:        "Message",
			TypeName:    "string",
			Description: "result message",
			Tags: map[string]string{
				"json": "msg",
			},
			TypeSpec: NewBasicType("string"),
		},
		{
			Name:        "Data",
			TypeName:    respData.TypeName(),
			Description: "result data",
			Tags: map[string]string{
				"json": "data",
			},
			TypeSpec: respData,
		},
	}
	return
}

// api 文件结构
type ApiFile struct {
	SourceFile          string `json:"source_file" yaml:"source_file" validate:"required"`                 // 源码
	PackageName         string `json:"api_handler_package" yaml:"api_handler_package" validate:"required"` // handler 所在的包名
	PackageExportedPath string `json:"package_exported" yaml:"package_exported"`                           // 被别人引入的path
	PackageDir          string `json:"package_dir" yaml:"package_dir"`                                     // 包所在的路径
	PackageRelAlias     string `json:"package_rel_alias" yaml:"package_rel_alias"`                         // 包昵称

	Summary     string   `json:"summary" yaml:"summary"`
	Description string   `json:"description" yaml:"description"`
	Tags        []string `json:"tags" yaml:"tags"`
	Deprecated  bool     `json:"deprecated" yaml:"deprecated"`
}

func NewApiFile() *ApiFile {
	return &ApiFile{
		Tags: make([]string, 0),
	}
}

func (m *ApiFile) MergeInfoFromCommentTags(tags *CommentTags) {
	if m.Summary == "" {
		m.Summary = tags.Summary
	}

	if m.Description == "" {
		m.Description = tags.Description
	}

	if len(tags.LineTagDocTags) > 0 {
		if m.Tags == nil {
			m.Tags = make([]string, 0)
		}
		m.Tags = append(m.Tags, tags.LineTagDocTags...)
	}

	if tags.Deprecated {
		m.Deprecated = true
	}
}
