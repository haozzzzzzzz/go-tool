package parse

import (
	"github.com/haozzzzzzzz/go-rapid-development/utils/ujson"
	"github.com/haozzzzzzzz/go-tool/common/source"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"sort"
)

type WsMsgIdOutput struct {
	Name  string       `json:"name" yaml:"name"`
	Value string       `json:"value" yaml:"value"`
	Title string       `json:"title" yaml:"title"`
	Doc   string       `json:"doc" yaml:"doc"`
	IType source.IType `json:"itype" yaml:"itype"`
}

type WsMsgCommonOutput struct {
	Title string       `json:"title" yaml:"title"`
	Doc   string       `json:"doc" yaml:"doc"`
	IType source.IType `json:"itype" yaml:"itype"`
}

type WsMsgBodyOutput struct {
	MsgId string       `json:"msg_id" yaml:"msg_id"`
	Title string       `json:"title" yaml:"title"`
	Doc   string       `json:"doc" yaml:"doc"`
	IType source.IType `json:"itype" yaml:"itype"`
}

// ws types output structure
type WsTypesOutput struct {
	SpecVersion    string               `json:"spec_version"`
	MsgIds         []*WsMsgIdOutput     `json:"-" yaml:"-"`
	UpMsgIds       []*WsMsgIdOutput     `json:"up_msg_ids" yaml:"up_msg_ids"`
	DownMsgIds     []*WsMsgIdOutput     `json:"down_msg_ids" yaml:"down_msg_ids"`
	UpMsgCommons   []*WsMsgCommonOutput `json:"up_msg_commons" yaml:"up_msg_commons"`
	UpMsgBodys     []*WsMsgBodyOutput   `json:"up_msg_bodys" yaml:"up_msg_bodys"`
	DownMsgCommons []*WsMsgCommonOutput `json:"down_msg_commons" yaml:"down_msg_commons"`
	DownMsgBodys   []*WsMsgBodyOutput   `json:"down_msg_bodys" yaml:"down_msg_bodys"`
}

const LatestWsTypeOutputSpecVersion string = "v0.1.0"

func NewWsTypesOutput() (output *WsTypesOutput) {
	output = &WsTypesOutput{
		SpecVersion:    LatestWsTypeOutputSpecVersion,
		MsgIds:         []*WsMsgIdOutput{},
		UpMsgIds:       []*WsMsgIdOutput{},
		DownMsgIds:     []*WsMsgIdOutput{},
		UpMsgCommons:   []*WsMsgCommonOutput{},
		UpMsgBodys:     []*WsMsgBodyOutput{},
		DownMsgCommons: []*WsMsgCommonOutput{},
		DownMsgBodys:   []*WsMsgBodyOutput{},
	}
	return
}

func (m *WsTypesOutput) SortOut() {
	// sort msg ids
	msgIdMap := make(map[string]*WsMsgIdOutput)
	strMsgIds := make([]string, 0)
	for _, msgId := range m.MsgIds {
		msgIdMap[msgId.Value] = msgId
		strMsgIds = append(strMsgIds, msgId.Value)
	}

	sort.Strings(strMsgIds)

	upMsgBodyMap := make(map[string]*WsMsgBodyOutput)
	for _, msgBody := range m.UpMsgBodys {
		upMsgBodyMap[msgBody.MsgId] = msgBody
	}

	downMsgBodyMap := make(map[string]*WsMsgBodyOutput)
	for _, msgBody := range m.DownMsgBodys {
		downMsgBodyMap[msgBody.MsgId] = msgBody
	}

	msgIds := make([]*WsMsgIdOutput, 0)
	upMsgBodys := make([]*WsMsgBodyOutput, 0)
	downMsgBodys := make([]*WsMsgBodyOutput, 0)
	for _, strMsgId := range strMsgIds {
		msgId := msgIdMap[strMsgId]
		msgIds = append(msgIds, msgId)

		upBody, ok := upMsgBodyMap[strMsgId]
		if ok {
			upMsgBodys = append(upMsgBodys, upBody)
			m.UpMsgIds = append(m.UpMsgIds, msgId)
		}

		downBody, ok := downMsgBodyMap[strMsgId]
		if ok {
			downMsgBodys = append(downMsgBodys, downBody)
			m.DownMsgIds = append(m.DownMsgIds, msgId)
		}
	}

	m.MsgIds = msgIds
	m.UpMsgBodys = upMsgBodys
	m.DownMsgBodys = downMsgBodys
}

func (m *WsTypesOutput) Json() (bObj []byte, err error) {
	bObj, err = ujson.MarshalPretty(m)
	if err != nil {
		logrus.Errorf("json marshal ws types output failed. error: %s", err)
		return
	}

	return
}

func (m *WsTypesOutput) Yaml() (bObj []byte, err error) {
	bObj, err = yaml.Marshal(m)
	if err != nil {
		logrus.Errorf("yaml marshal ws types output failed. error: %s", err)
		return
	}
	return
}
