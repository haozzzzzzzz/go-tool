package source

import (
	"github.com/haozzzzzzzz/go-rapid-development/v2/utils/uerrors"
	"strings"
)

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

func (m *BasicType) TypeName() string {
	return m.Name
}

func (m *BasicType) TypeDescription() string {
	return m.Description
}

func NewBasicType(name string) *BasicType {
	return &BasicType{
		Name:      name,
		TypeClass: TypeClassBasicType,
	}
}

type TagsMap map[string]string

// ApiDocTagAttributes field api_doc tag
// `api_doc:""`
type ApiDocTagAttributes struct {
	Skip    bool   `json:"skip"`     // wheter skip generate field doc
	StrType string `json:"str_type"` // parse field type as spec str_type
}

func (m TagsMap) ApiDoc() (
	parts []string,
	attributes ApiDocTagAttributes,
) {
	parts = make([]string, 0)
	if m == nil {
		return
	}

	str := m["api_doc"]
	if str == "" {
		return
	}

	tempParts := strings.Split(str, ",")
	for _, part := range tempParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		parts = append(parts, part)
		switch {
		case part == "skip":
			attributes.Skip = true
		case strings.HasPrefix(part, "type="):
			attributes.StrType = strings.TrimLeft(part, "type=")
		}
	}

	return
}

// HasApiDocSkip whether skip generate field's doc
func (m TagsMap) HasApiDocSkip() (isSkip bool) {
	_, attributes := m.ApiDoc()
	isSkip = attributes.Skip
	return
}

// Field struct field
type Field struct {
	Name     string  `json:"name" yaml:"name"` // field name
	TypeName string  `json:"type_name" yaml:"type_name"`
	Tags     TagsMap `json:"tags" yaml:"tags"`
	TypeSpec IType   `json:"type_spec" yaml:"type_spec"`

	Exported bool `json:"exported" yaml:"exported"` // 是否是公开访问的。embedded类型为false

	/*
		是否是嵌入的属性。
		type A struct {
			B // B是A的embedded field
		}
	*/
	Embedded    bool   `json:"embedded" yaml:"embedded"`
	Description string `json:"description" yaml:"description"`
}

func (m *Field) TagJsonOrName() (name string) {
	name = m.TagJson()
	if name == "" {
		name = m.Name
	}
	return
}

func (m *Field) TagJson() (name string) {
	strJson := m.Tags["json"]
	if strJson == "" {
		return
	}

	jsonParts := strings.Split(strJson, ",")
	if len(jsonParts) == 0 {
		return
	}

	name = strings.TrimSpace(jsonParts[0])
	return
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

// struct type
type StructType struct {
	TypeClass   string            `json:"type_class" yaml:"type_class"`
	Name        string            `json:"name" yaml:"name"`
	Embedded    []*Field          `json:"embedded" yaml:"embedded"`
	Fields      []*Field          `json:"fields" yaml:"fields"`
	mField      map[string]*Field `json:"-" yaml:"-"`
	Description string            `json:"description" yaml:"description"`

	Underlying *StructType `json:"-" yaml:"-"` // 底层类。如果一个类被重新定义, type A B后，我们认为A、B类底层同属于B
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

func (m *StructType) AddEmbedded(fields ...*Field) {
	m.Embedded = append(m.Embedded, fields...)
	return
}

func (m *StructType) IsEmpty() bool {
	return len(m.Fields) == 0
}

func (m *StructType) Copy() *StructType {
	clone := *m

	if clone.Underlying != nil {
		clone.Underlying = m
	}

	return &clone
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

// ArrayType array
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
