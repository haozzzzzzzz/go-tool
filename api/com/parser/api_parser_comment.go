package parser

import (
	"bufio"
	"encoding/json"
	"github.com/haozzzzzzzz/go-tool/common/source"
	"io"
	"regexp"

	"reflect"

	"strings"

	"fmt"

	"github.com/haozzzzzzzz/go-rapid-development/v2/utils/uerrors"
	"github.com/sirupsen/logrus"
)

const BlockTagKeyCommonStart = "api_doc_common_start"
const BlockTagKeyCommonEnd = "api_doc_common_end"
const BlockTagKeyStart = "api_doc_start"
const BlockTagKeyEnd = "api_doc_end"

const LineTagKeyDocRoute = "api_doc_route"
const LineTagKeyDocHttpMethod = "api_doc_http_method"
const LineTagKeyDocRelativePaths = "api_doc_relative_paths"
const LineTagKeyDocTags = "api_doc_tags"

func ParseApisFromPkgCommentText(
	commentText string,
) (
	commonParams []*ApiItemParams,
	apis []*ApiItem,
	err error,
) {
	commonApis, err := parseApisFromCommentText(
		commentText,
		"@"+BlockTagKeyCommonStart,
		"@"+BlockTagKeyCommonEnd,
	)
	if nil != err {
		logrus.Errorf("parse api common params from comment failed. error: %s.", err)
		return
	}

	commonParams = make([]*ApiItemParams, 0)
	for _, commonApi := range commonApis {
		commonParams = append(commonParams, &commonApi.ApiItemParams)
	}

	apis, err = parseApisFromCommentText(
		commentText,
		"@"+BlockTagKeyStart,
		"@"+BlockTagKeyEnd,
	)
	if nil != err {
		logrus.Errorf("parse apis from comment failed. error: %s.", err)
		return
	}

	for _, api := range apis {
		if api.HttpMethod == "" || len(api.RelativePaths) == 0 {
			logrus.Warnf("found empty api doc comment, require http_method and relative_paths")
			return
		}
	}

	return
}

func parseApisFromCommentText(
	commentText string,
	startTag string,
	endTag string,
) (apis []*ApiItem, err error) {
	apis = make([]*ApiItem, 0)
	strReg := fmt.Sprintf("(?si:%s(.*?)%s)", startTag, endTag)
	docReg, err := regexp.Compile(strReg)
	if nil != err {
		logrus.Errorf("reg compile pkg api comment text failed. error: %s.", err)
		return
	}

	arrStrs := docReg.FindAllStringSubmatch(commentText, -1)
	strJsons := make([]string, 0)
	for _, strs := range arrStrs {
		strJsons = append(strJsons, strs[1])
	}

	for _, strJson := range strJsons {
		tempApiItem, errParse := parseCommentTextToApi(strJson)
		err = errParse
		if nil != err {
			logrus.Errorf("parse comment to to api failed. %s, error: %s.", strJson, err)
			return
		}

		if tempApiItem == nil {
			continue
		}

		apis = append(apis, tempApiItem)
	}

	return
}

/**
@api_doc_start
{
	"http_method": "GET",
	"relative_paths": ["/hello_world"],
	"query_data": {
		"name": "姓名|string|required"
	},
	"post_data": {
		"location": "地址|string|required"
	},
	"resp_data": {
	    "a": "a|int",
	    "b": "b|int",
	    "c": {
			"d": "d|string"
		},
		"__c": "c|object",
		"f": [
			"string"
		],
		"__f": "f|object|required",
		"g": [
			{
				"h": "h|string|required"
			}
		],
		"__g": "g|array|required"
	}
}
@api_doc_end
*/
type CommentTextApi struct {
	HttpMethod    string   `json:"http_method"`
	RelativePaths []string `json:"relative_paths"`

	UriData    map[string]interface{} `json:"uri_data"`
	HeaderData map[string]interface{} `json:"header_data"`
	QueryData  map[string]interface{} `json:"query_data"`
	PostData   map[string]interface{} `json:"post_data"`
	RespData   map[string]interface{} `json:"resp_data"`
}

func parseCommentTextToApi(
	text string,
) (api *ApiItem, err error) {
	comApi := &CommentTextApi{}
	err = json.Unmarshal([]byte(text), comApi)
	if nil != err {
		logrus.Errorf("unmarshal api failed. error: %s.", err)
		return
	}

	api = &ApiItem{
		HttpMethod:    comApi.HttpMethod,
		RelativePaths: comApi.RelativePaths,
	}

	api.UriData, err = commentApiRequestDataToStructType(comApi.UriData)
	if nil != err {
		logrus.Errorf("comment text api path data to struct type failed. error: %s.", err)
		return
	}

	api.HeaderData, err = commentApiRequestDataToStructType(comApi.HeaderData)
	if nil != err {
		logrus.Errorf("comment text api header data to struct type failed. error: %s.", err)
		return
	}

	api.QueryData, err = commentApiRequestDataToStructType(comApi.QueryData)
	if nil != err {
		logrus.Errorf("comment text api query data to struct type failed. error: %s.", err)
		return
	}

	postData, err := commentApiRequestDataToStructType(comApi.PostData)
	if nil != err {
		logrus.Errorf("comment text api post data to struct type failed. error: %s.", err)
		return
	}
	if postData != nil {
		api.PostData = postData
	}

	respData, err := commentApiRequestDataToStructType(comApi.RespData)
	if nil != err {
		logrus.Errorf("comment text api resp data to struct type failed. error: %s.", err)
		return
	}
	if respData != nil {
		api.RespData = respData
	}

	return
}

func commentApiRequestDataToStructType(
	mapData map[string]interface{},
) (structType *source.StructType, err error) {
	if mapData == nil {
		return
	}

	structType = source.NewStructType()
	for key, typeDesc := range mapData {
		if strings.HasPrefix(key, "__") {
			continue
		}

		keyDesc := mapData[fmt.Sprintf("__%s", key)] // 如果是嵌套类型，则会有一个__key描述这个field在当前struct的属性
		strKeyDesc, _ := keyDesc.(string)
		field, errField := commentApiRequestDataFieldDesc(key, typeDesc, strKeyDesc)
		err = errField
		if nil != err {
			logrus.Errorf("get struct field failed. key: %s, type_desc: %#v error: %s.", key, typeDesc, err)
			return
		}

		err = structType.AddFields(field)
		if nil != err {
			logrus.Errorf("struct type add fields failed. error: %s.", err)
			return
		}
	}
	return
}

func commentApiRequestDataFieldDesc(key string, fieldTypeDesc interface{}, slaveFieldDesc string) (field *source.Field, err error) {
	if fieldTypeDesc == nil {
		return
	}

	field = source.NewField()
	field.Name = key

	strFieldTypeDesc, ok := fieldTypeDesc.(string)
	if !ok {
		strFieldTypeDesc = slaveFieldDesc
	}

	if strFieldTypeDesc != "" {
		vals := parseFieldTypeDescParts(strFieldTypeDesc)
		field.Description = vals[0]
		field.TypeName = vals[1]
		field.Tags["json"] = key
		field.Tags["binding"] = vals[2]
	}

	field.TypeSpec, err = commentApiRequestDataJsonDescToIType(fieldTypeDesc)
	if nil != err {
		logrus.Errorf("parse field type spec failed. error: %s.", err)
		return
	}

	return
}

func parseFieldTypeDescParts(strFieldTypeDesc string) (vals [3]string) {
	vals = [3]string{} // description, type, tags
	splitDefs := strings.Split(strFieldTypeDesc, "|")
	for i := 0; i < 3 && i < len(splitDefs); i++ {
		vals[i] = splitDefs[i]
	}
	return
}

func commentApiRequestDataJsonDescToIType(
	typeDesc interface{},
) (itype source.IType, err error) {
	reflectType := reflect.TypeOf(typeDesc)
	switch reflectType.Kind() {
	case reflect.String:
		strTypeDes := typeDesc.(string)
		vals := parseFieldTypeDescParts(strTypeDes)
		itype = source.NewBasicType(vals[1])

	case reflect.Map:
		mTypeDesc, ok := typeDesc.(map[string]interface{})
		if !ok {
			logrus.Warnf("convert def to map type failed. typeDesc: %#v", typeDesc)
			return
		}

		itype, err = commentApiRequestDataToStructType(mTypeDesc)
		if nil != err {
			logrus.Errorf("field map type desc to struct type failed. error: %s.", err)
			return
		}

	case reflect.Slice:
		sliceTypeDesc, ok := typeDesc.([]interface{})
		if !ok {
			logrus.Warnf("convert def to slice type failed. typeDesc: %#v", typeDesc)
			return
		}

		sliceType := source.NewArrayType()
		if len(sliceTypeDesc) > 0 {
			sliceType.EltSpec, err = commentApiRequestDataJsonDescToIType(sliceTypeDesc[0])
			if nil != err {
				logrus.Errorf("parse slice type elt spec failed. error: %s.", err)
				return
			}
		}

		itype = sliceType

	default:
		err = uerrors.Newf("unsupported type: %#v", typeDesc)

	}
	return
}

// 注释标签
type CommentTags struct {
	Summary     string // 非tag的注释第一行是summary，其余是description
	Description string
	Deprecated  bool

	LineTagDocRoute string

	LineTagDocHttpMethod    string
	LineTagDocRelativePaths []string
	LineTagDocTags          []string
}

func NewCommentTags() *CommentTags {
	return &CommentTags{
		LineTagDocRelativePaths: make([]string, 0),
	}
}

func ParseCommentTags(text string) (tags *CommentTags, err error) {
	tags = NewCommentTags()

	// 去掉头部多余的*
	text = strings.TrimSpace(text)
	text = strings.Trim(text, "*")
	text = strings.TrimSpace(text)

	// read tags
	strBuf := bufio.NewReader(strings.NewReader(text))
	docReg, errC := regexp.Compile(`(?i:\@(.*?)\:)`)
	err = errC
	if nil != err || docReg == nil {
		logrus.Errorf("compile line tag regexp failed. error: %s.", err)
		return
	}

	noTagsText := ""
	for {
		bLine, _, errR := strBuf.ReadLine()
		err = errR
		if nil != err && err != io.EOF {
			logrus.Errorf("read description failed. error: %s.", err)
			return
		}

		if err == io.EOF {
			err = nil
			break
		}

		line := string(bLine)
		matcheds := docReg.FindAllStringSubmatch(line, 1)
		if len(matcheds) != 1 || len(matcheds[0]) != 2 {
			if noTagsText == "" {
				noTagsText = line
			} else {
				noTagsText += "\n" + line
			}
			continue
		}

		var setTagsRelativePaths = func(strRelativePaths string) {
			relPaths := strings.Split(strRelativePaths, ",")
			for _, relPath := range relPaths {
				relPath = strings.TrimSpace(relPath)
				if relPath == "" {
					continue
				}

				tags.LineTagDocRelativePaths = append(tags.LineTagDocRelativePaths, relPath)
			}
		}

		var setTagsTags = func(strDocTags string) {
			docTags := strings.Split(strDocTags, ",")
			for _, docTag := range docTags {
				docTag = strings.TrimSpace(docTag)
				if docTag == "" {
					continue
				}

				tags.LineTagDocTags = append(tags.LineTagDocTags, docTag)
			}
		}

		rawMatchedString := matcheds[0][0]
		tagKey := strings.TrimSpace(matcheds[0][1])
		switch tagKey {
		case LineTagKeyDocHttpMethod:
			tags.LineTagDocHttpMethod = strings.TrimSpace(strings.Replace(line, rawMatchedString, "", 1))

		case LineTagKeyDocRelativePaths:
			strRelativePaths := strings.TrimSpace(strings.Replace(line, rawMatchedString, "", 1))
			setTagsRelativePaths(strRelativePaths)

		case LineTagKeyDocTags:
			strDocTags := strings.TrimSpace(strings.Replace(line, rawMatchedString, "", 1))
			setTagsTags(strDocTags)

		case LineTagKeyDocRoute:
			strRoute := strings.TrimSpace(strings.Replace(line, rawMatchedString, "", 1))
			strParts := strings.Split(strRoute, "|")
			lenParts := len(strParts)
			if lenParts > 0 {
				tags.LineTagDocHttpMethod = strings.TrimSpace(strParts[0])
			}

			if lenParts > 1 {
				setTagsRelativePaths(strParts[1])
			}

			if lenParts > 2 {
				setTagsTags(strParts[2])
			}

		default:
			logrus.Warnf("unsupported api comment line tag: %s", tagKey)

		}

	}

	// read summary description
	if noTagsText != "" {
		strBuf = bufio.NewReader(strings.NewReader(noTagsText))
		bLine, _, errRead := strBuf.ReadLine()
		err = errRead
		if nil != err {
			logrus.Errorf("read api comment first line failed. error: %s.", err)
			return
		}

		tags.Summary = string(bLine)
		tags.Description = strings.Replace(noTagsText, tags.Summary, "", 1)
		tags.Summary = strings.TrimSpace(tags.Summary)
		tags.Description = strings.TrimSpace(tags.Description)

		strBuf = bufio.NewReader(strings.NewReader(noTagsText))
		if strings.Contains(noTagsText, "Deprecated") {
			tags.Deprecated = true
		}
	}

	return
}
