package parse

import (
	"github.com/haozzzzzzzz/go-rapid-development/utils/ujson"
	"github.com/haozzzzzzzz/go-tool/common/source"
	"github.com/sirupsen/logrus"
	"strings"
)

// tag key类型
type WsTagKeyType string

const WsTagKeyUpCommon WsTagKeyType = "ws_doc_up_common"     // 上行消息公共参数
const WsTagKeyDownCommon WsTagKeyType = "ws_doc_down_common" // 下行消息公共参数
const WsTagKeyCommon WsTagKeyType = "ws_doc_common"          // 上下行消息公共参数
const WsTagKeyUpBody WsTagKeyType = "ws_doc_up_body"         // 上行消息消息体
const WsTagKeyDownBody WsTagKeyType = "ws_doc_down_body"     // 下行消息消息体
const WsTagKeyBody WsTagKeyType = "ws_doc_body"              // 上下行消息消息体

type WsTagBody struct {
	TagKey WsTagKeyType `json:"tag_key"`
	MsgId  string       `json:"msg_id"` // 消息Id
}

// <0:msg_id>|<...>
func NewWsTagBody(tagKey WsTagKeyType, tagValue string) (body *WsTagBody) {
	body = &WsTagBody{
		TagKey: tagKey,
	}

	valParts := strings.Split(tagValue, "|")
	lenParts := len(valParts)
	if lenParts > 0 {
		body.MsgId = strings.TrimSpace(valParts[0])
	}
	return
}

type WsTagDoc struct {
	Lines []string `json:"lines"`
}

type WsTag struct {
	Valid         bool           `json:"valid"` // 是否包含WsTag
	TagKeys       []WsTagKeyType `json:"tag_keys"`
	HasUpCommon   bool           `json:"has_up_common"`
	HasDownCommon bool           `json:"has_down_common"`
	UpBodys       []*WsTagBody   `json:"up_bodys"`
	DownBodys     []*WsTagBody   `json:"down_bodys"`
	Doc           *WsTagDoc      `json:"doc"`
}

func NewWsTag() (tag *WsTag) {
	tag = &WsTag{
		Valid:     false,
		TagKeys:   make([]WsTagKeyType, 0),
		UpBodys:   make([]*WsTagBody, 0),
		DownBodys: make([]*WsTagBody, 0),
		Doc:       nil,
	}
	return
}

func (m *WsTag) String() (str string) {
	bObj, err := ujson.MarshalPretty(m)
	if err != nil {
		logrus.Errorf("marshal ws tag failed. error: %s", err)
		return
	}
	str = string(bObj)
	return
}

func (m *WsTag) parseUpCommon(tagKey WsTagKeyType, tagValue string) (err error) {
	m.HasUpCommon = true
	return
}

func (m *WsTag) parseDownCommon(tagKey WsTagKeyType, tagValue string) (err error) {
	m.HasDownCommon = true
	return
}

func (m *WsTag) parseCommon(tagKey WsTagKeyType, tagValue string) (err error) {
	err = m.parseUpCommon(tagKey, tagValue)
	if err != nil {
		logrus.Errorf("parse common for up failed. error: %s", err)
		return
	}

	err = m.parseDownCommon(tagKey, tagValue)
	if err != nil {
		logrus.Errorf("parse common for down failed. error: %s", err)
		return
	}
	return
}

func (m *WsTag) parseUpBody(tagKey WsTagKeyType, tagValue string) (err error) {
	body := NewWsTagBody(tagKey, tagValue)
	m.UpBodys = append(m.UpBodys, body)
	return
}

func (m *WsTag) parseDownBody(tagKey WsTagKeyType, tagValue string) (err error) {
	body := NewWsTagBody(tagKey, tagValue)
	m.DownBodys = append(m.DownBodys, body)
	return
}

func (m *WsTag) parseBody(tagKey WsTagKeyType, tagValue string) (err error) {
	err = m.parseUpBody(tagKey, tagValue)
	if err != nil {
		logrus.Errorf("parse body for up failed. error: %s", err)
		return
	}

	err = m.parseDownBody(tagKey, tagValue)
	if err != nil {
		logrus.Errorf("parse body for down failed. error: %s", err)
		return
	}
	return
}

type TagParseHandler func(tagKey WsTagKeyType, tagValue string) (err error)

func (m *WsTag) parseTag(tagKey WsTagKeyType, tagValue string) (err error) {
	tagParsers := map[WsTagKeyType]TagParseHandler{
		WsTagKeyUpCommon:   m.parseUpCommon,
		WsTagKeyDownCommon: m.parseDownCommon,
		WsTagKeyCommon:     m.parseCommon,
		WsTagKeyUpBody:     m.parseUpBody,
		WsTagKeyDownBody:   m.parseDownBody,
		WsTagKeyBody:       m.parseBody,
	}

	tagParser, ok := tagParsers[tagKey]
	if !ok {
		return
	}

	err = tagParser(tagKey, tagValue)
	if err != nil {
		logrus.Errorf("parse tag %s failed. error: %s", tagKey, err)
		return
	}

	m.TagKeys = append(m.TagKeys, tagKey)

	m.Valid = true
	return

}

func WsTagFromCommentText(text string) (
	wsTag *WsTag,
	err error,
) {
	commentText, err := source.NewCommentText(text)
	if err != nil {
		logrus.Errorf("new comment text failed. error: %s", err)
		return
	}

	wsTag = NewWsTag()

	for tagKey, tagValue := range commentText.Tags {
		err = wsTag.parseTag(WsTagKeyType(tagKey), tagValue)
		if err != nil {
			logrus.Errorf("parse tag failed. error: %s", err)
			return
		}
	}

	if wsTag.Valid {
		wsTag.Doc = &WsTagDoc{
			Lines: commentText.Lines,
		}
	}

	return
}
