package parse

import (
	"github.com/haozzzzzzzz/go-tool/common/source"
	"github.com/sirupsen/logrus"
)

type UpDownMsgs struct {
	Common       source.IType            `json:"common" yaml:"common"`
	MsgIdBodyMap map[string]source.IType `json:"msg_id_body_map" yaml:"msg_id_body_map"`
}

func NewUpDownMsgs() *UpDownMsgs {
	return &UpDownMsgs{
		MsgIdBodyMap: map[string]source.IType{},
	}
}

type WsTypes struct {
	UpMsgs   *UpDownMsgs `json:"up_msgs" yaml:"up_msgs"`     // 上行消息
	DownMsgs *UpDownMsgs `json:"down_msgs" yaml:"down_msgs"` // 下行消息
}

func NewWsTypes() *WsTypes {
	return &WsTypes{
		UpMsgs:   NewUpDownMsgs(),
		DownMsgs: NewUpDownMsgs(),
	}
}

type WsTypesParser struct {
	parser  *source.TypesParser
	WsTypes *WsTypes
}

func NewWsTypesParser(
	rootDir string,
) *WsTypesParser {
	return &WsTypesParser{
		parser:  source.NewTypesParser(rootDir),
		WsTypes: NewWsTypes(),
	}
}

// 解析含有ws标签的类型
func (m *WsTypesParser) ParseWsTypes() (err error) {
	err = m.parser.Parse(m.parseWsTypeFilter)
	if err != nil {
		logrus.Errorf("parse source types failed. error: %s", err)
		return
	}
	return
}

func (m *WsTypesParser) parseWsTypeFilter() (match bool) {
	// TODO
	return
}
