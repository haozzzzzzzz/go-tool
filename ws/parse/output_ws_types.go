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
	MsgIds         []*WsMsgIdOutput     `json:"msg_ids" yaml:"msg_ids"`
	UpMsgCommons   []*WsMsgCommonOutput `json:"up_msg_commons" yaml:"up_msg_commons"`
	UpMsgBodys     []*WsMsgBodyOutput   `json:"up_msg_bodys" yaml:"up_msg_bodys"`
	DownMsgCommons []*WsMsgCommonOutput `json:"down_msg_commons" yaml:"down_msg_commons"`
	DownMsgBodys   []*WsMsgBodyOutput   `json:"down_msg_bodys" yaml:"down_msg_bodys"`
}

func NewWsTypesOutput() (output *WsTypesOutput) {
	output = &WsTypesOutput{
		MsgIds:         []*WsMsgIdOutput{},
		UpMsgCommons:   []*WsMsgCommonOutput{},
		UpMsgBodys:     []*WsMsgBodyOutput{},
		DownMsgCommons: []*WsMsgCommonOutput{},
		DownMsgBodys:   []*WsMsgBodyOutput{},
	}
	return
}

func (m *WsTypesOutput) Sort() {
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
		msgIds = append(msgIds, msgIdMap[strMsgId])

		upBody, ok := upMsgBodyMap[strMsgId]
		if ok {
			upMsgBodys = append(upMsgBodys, upBody)
		}

		downBody, ok := downMsgBodyMap[strMsgId]
		if ok {
			downMsgBodys = append(downMsgBodys, downBody)
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
