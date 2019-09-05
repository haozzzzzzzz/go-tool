package parser

import (
	"fmt"
	"github.com/haozzzzzzzz/go-rapid-development/utils/uerrors"
	"github.com/sirupsen/logrus"
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
}

// 类型分类
const TypeClassBasicType = "basic"
const TypeClassStructType = "struct"
const TypeClassMapType = "map"
const TypeClassArrayType = "array"
const TypeClassInterfaceType = "interface"

// 标准类型
type BasicType struct {
	TypeClass string `json:"type_class" yaml:"type_class"`
	Name      string `json:"name" yaml:"name"`
}

func (m BasicType) TypeName() string {
	return string(m.Name)
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

func (m *StructType) AddFields(fields ...*Field) (err error) {
	for _, field := range fields {
		if field.Name == "" {
			err = uerrors.Newf("struct field require name")
			return
		}

		_, ok := m.mField[field.TagJsonOrName()]
		if ok {
			err = uerrors.Newf("has conflict field name")
			return
		}

		m.Fields = append(m.Fields, field)

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
	TypeClass string `json:"type_class" yaml:"type_class"`
	Name      string `json:"name" yaml:"name"`
	KeySpec   IType  `json:"key" yaml:"key"`
	ValueSpec IType  `json:"value_spec" yaml:"value_spec"`
}

func (m *MapType) TypeName() string {
	return m.Name
}

func NewMapType() *MapType {
	return &MapType{
		TypeClass: TypeClassMapType,
		Name:      TypeClassMapType,
	}
}

// array
type ArrayType struct {
	TypeClass string `json:"type_class" yaml:"type_class"`
	Name      string `json:"name" yaml:"name"`
	EltName   string `json:"elt_name" yaml:"elt_name"`
	EltSpec   IType  `json:"elt_spec" yaml:"elt_spec"`
}

func (m *ArrayType) TypeName() string {
	return m.Name
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

type ApiItemParams struct {
	HeaderData *StructType `json:"header_data" yaml:"header_data"`
	UriData    *StructType `json:"uri_data" yaml:"uri_data"`
	QueryData  *StructType `json:"query_data" yaml:"query_data"`
	PostData   *StructType `json:"post_data" yaml:"post_data"`
	RespData   *StructType `json:"response_data" yaml:"response_data"`
}

func NewApiItemParams() *ApiItemParams {
	return &ApiItemParams{
		HeaderData: NewStructType(),
		UriData:    NewStructType(),
		QueryData:  NewStructType(),
		PostData:   NewStructType(),
		RespData:   NewStructType(),
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

		if item.PostData != nil && len(item.PostData.Fields) > 0 {
			if m.PostData == nil {
				m.PostData = NewStructType()
			}

			err = m.PostData.AddFields(item.PostData.Fields...)
			if nil != err {
				logrus.Errorf("add post data fields failed. error: %s.", err)
				return
			}
		}

		if item.RespData != nil && len(item.RespData.Fields) > 0 {
			if m.RespData == nil {
				m.RespData = NewStructType()
			}

			err = m.RespData.AddFields(item.RespData.Fields...)
			if nil != err {
				logrus.Errorf("add resp data fields failed. error: %s.", err)
				return
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

type ApiItem struct {
	ApiItemParams

	ApiHandlerFunc string `validate:"required" json:"api_handler_func" yaml:"api_handler_func"` // handler 函数名
	SourceFile     string `validate:"required" json:"source_file" yaml:"source_file"`           // 源码

	PackageName         string `validate:"required" json:"api_handler_package" yaml:"api_handler_package"` // handler 所在的包名
	PackageExportedPath string `json:"package_exported" yaml:"package_exported"`                           // 被别人引入的path
	PackageDir          string `json:"package_dir" yaml:"package_dir"`                                     // 包所在的路径
	PackageRelAlias     string `json:"package_rel_alias" yaml:"package_rel_alias"`                         // 包昵称

	HttpMethod    string   `validate:"required" json:"http_method" yaml:"http_method"`
	RelativePaths []string `validate:"required" json:"relative_paths" yaml:"relative_paths"`

	Summary     string `json:"summary" yaml:"summary"`
	Description string `json:"description" yaml:"description"`
}

func (m *ApiItem) PackageFuncName() string {
	return fmt.Sprintf("%s.%s", m.PackageRelAlias, m.ApiHandlerFunc)
}

// 成功返回结构
//type Response struct {
//	ReturnCode uint32      `json:"ret"`
//	Message    string      `json:"msg"`
//	Data       interface{} `json:"data"`
//}
func SuccessResponseStructType(
	respData *StructType,
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
