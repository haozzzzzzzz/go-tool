package parser

import (
	"fmt"
	"github.com/haozzzzzzzz/go-tool/common/source"
	"github.com/sirupsen/logrus"
	"reflect"
)

type ApiItemParams struct {
	HeaderData *source.StructType `json:"header_data" yaml:"header_data"`
	UriData    *source.StructType `json:"uri_data" yaml:"uri_data"`
	QueryData  *source.StructType `json:"query_data" yaml:"query_data"`
	PostData   source.IType       `json:"post_data" yaml:"post_data"`
	RespData   source.IType       `json:"response_data" yaml:"response_data"`
}

func NewApiItemParams() *ApiItemParams {
	return &ApiItemParams{
		HeaderData: source.NewStructType(),
		UriData:    source.NewStructType(),
		QueryData:  source.NewStructType(),
		PostData:   source.NewStructType(), // default
		RespData:   source.NewStructType(), // default
	}
}

func (m *ApiItemParams) MergeApiItemParams(items ...*ApiItemParams) (err error) {
	for _, item := range items {
		if item == nil {
			continue
		}

		if item.HeaderData != nil && len(item.HeaderData.Fields) > 0 {
			if m.HeaderData == nil {
				m.HeaderData = source.NewStructType()
			}

			err = m.HeaderData.AddFields(item.HeaderData.Fields...)
			if nil != err {
				logrus.Errorf("add header data fields failed. error: %s.", err)
				return
			}
		}

		if item.UriData != nil && len(item.UriData.Fields) > 0 {
			if m.UriData == nil {
				m.UriData = source.NewStructType()
			}

			err = m.UriData.AddFields(item.UriData.Fields...)
			if nil != err {
				logrus.Errorf("add uri data fields failed. error: %s.", err)
				return
			}
		}

		if item.QueryData != nil && len(item.QueryData.Fields) > 0 {
			if m.QueryData == nil {
				m.QueryData = source.NewStructType()
			}

			err = m.QueryData.AddFields(item.QueryData.Fields...)
			if nil != err {
				logrus.Errorf("add query data fields failed. error: %s.", err)
				return
			}
		}

		if item.PostData != nil {
			switch itemPostData := item.PostData.(type) {
			case *source.StructType:
				if itemPostData.IsEmpty() {
					break
				}

				if m.PostData == nil {
					m.PostData = source.NewStructType()
				}

				mPostStruct, ok := m.PostData.(*source.StructType)
				if !ok {
					logrus.Warnf("can not merge post data, because of different type. %s %s", reflect.TypeOf(item.PostData), reflect.TypeOf(m.PostData))
					break
				}

				err = mPostStruct.AddFields(itemPostData.Fields...)
				if nil != err {
					logrus.Errorf("add post data fields failed. error: %s.", err)
					return
				}

			}
		}

		if item.RespData != nil {
			if m.RespData == nil {
				m.RespData = source.NewStructType()
			}

			switch item.RespData.(type) {
			case *source.StructType: // only merge struct type
				itemRespStruct := item.RespData.(*source.StructType)
				if itemRespStruct == nil {
					break
				}

				mRespStruct, ok := m.RespData.(*source.StructType)
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

func NewApiItem() *ApiItem {
	return &ApiItem{
		RelativePaths: []string{},
		Tags:          []string{},
	}
}

func (m *ApiItem) PackageFuncName() string {
	return fmt.Sprintf("%s.%s", m.ApiFile.PackageRelAlias, m.ApiHandlerFunc)
}

func (m *ApiItem) ApiHandlerFuncPath() string {
	return fmt.Sprintf("%s:%s", m.ApiFile.SourceFile, m.ApiHandlerFunc)
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
	respData source.IType,
) (successResp *source.StructType) {
	successResp = source.NewStructType()
	successResp.Name = "SuccessResponse"
	successResp.Description = "api success response"
	successResp.Fields = []*source.Field{
		{
			Name:        "ReturnCode",
			TypeName:    "uint32",
			Description: "result code",
			Tags: map[string]string{
				"json": "ret",
			},
			Embedded: false,
			Exported: true,
			TypeSpec: source.NewBasicType("uint32"),
		},
		{
			Name:        "Message",
			TypeName:    "string",
			Description: "result message",
			Tags: map[string]string{
				"json": "msg",
			},
			Embedded: false,
			Exported: true,
			TypeSpec: source.NewBasicType("string"),
		},
		{
			Name:        "Data",
			TypeName:    respData.TypeName(),
			Description: "result data",
			Tags: map[string]string{
				"json": "data",
			},
			Embedded: false,
			Exported: true,
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
