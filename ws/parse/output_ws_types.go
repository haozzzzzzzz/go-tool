package parse

import (
	"github.com/haozzzzzzzz/go-rapid-development/v2/utils/ujson"
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
	SpecVersion     string                    `json:"spec_version"`
	MsgIds          []*WsMsgIdOutput          `json:"-" yaml:"-"`
	UpMsgIdValues   []string                  `json:"up_msg_id_values" yaml:"up_msg_id_values"`
	UpMsgIdMap      map[string]*WsMsgIdOutput `json:"up_msg_id_map" yaml:"up_msg_id_map"`
	DownMsgIdValues []string                  `json:"down_msg_id_values" yaml:"down_msg_id_values"`
	DownMsgIdMap    map[string]*WsMsgIdOutput `json:"down_msg_id_map" yaml:"down_msg_id_map"`
	UpMsgCommons    []*WsMsgCommonOutput      `json:"up_msg_commons" yaml:"up_msg_commons"`
	UpMsgBodys      []*WsMsgBodyOutput        `json:"up_msg_bodys" yaml:"up_msg_bodys"`
	DownMsgCommons  []*WsMsgCommonOutput      `json:"down_msg_commons" yaml:"down_msg_commons"`
	DownMsgBodys    []*WsMsgBodyOutput        `json:"down_msg_bodys" yaml:"down_msg_bodys"`
}

const LatestWsTypeOutputSpecVersion string = "v0.1.0"

func NewWsTypesOutput() (output *WsTypesOutput) {
	output = &WsTypesOutput{
		SpecVersion:     LatestWsTypeOutputSpecVersion,
		MsgIds:          []*WsMsgIdOutput{},
		UpMsgIdValues:   []string{},
		UpMsgIdMap:      map[string]*WsMsgIdOutput{},
		DownMsgIdValues: []string{},
		DownMsgIdMap:    map[string]*WsMsgIdOutput{},
		UpMsgCommons:    []*WsMsgCommonOutput{},
		UpMsgBodys:      []*WsMsgBodyOutput{},
		DownMsgCommons:  []*WsMsgCommonOutput{},
		DownMsgBodys:    []*WsMsgBodyOutput{},
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
			m.UpMsgIdValues = append(m.UpMsgIdValues, strMsgId)
			m.UpMsgIdMap[strMsgId] = msgId
		}

		downBody, ok := downMsgBodyMap[strMsgId]
		if ok {
			downMsgBodys = append(downMsgBodys, downBody)
			m.DownMsgIdValues = append(m.DownMsgIdValues, strMsgId)
			m.DownMsgIdMap[strMsgId] = msgId
		}
	}

	m.MsgIds = msgIds
	m.UpMsgBodys = upMsgBodys
	m.DownMsgBodys = downMsgBodys

	// commons
	upMsgCommons := make([]*WsMsgCommonOutput, 0)
	upMsgCommonMap := make(map[string]*WsMsgCommonOutput)
	upMsgCommonNames := make([]string, 0)

	downMsgCommons := make([]*WsMsgCommonOutput, 0)
	downMsgCommonMap := make(map[string]*WsMsgCommonOutput)
	downMsgCommonNames := make([]string, 0)

	for _, common := range m.UpMsgCommons {
		typeName := common.IType.TypeName()
		upMsgCommonMap[typeName] = common
		upMsgCommonNames = append(upMsgCommonNames, typeName)
	}

	for _, common := range m.DownMsgCommons {
		typeName := common.IType.TypeName()
		downMsgCommonMap[typeName] = common
		downMsgCommonNames = append(downMsgCommonNames, typeName)
	}

	sort.Strings(upMsgCommonNames)
	sort.Strings(downMsgCommonNames)

	for _, typeName := range upMsgCommonNames {
		upMsgCommons = append(upMsgCommons, upMsgCommonMap[typeName])
	}

	for _, typeName := range downMsgCommonNames {
		downMsgCommons = append(downMsgCommons, downMsgCommonMap[typeName])
	}

	m.UpMsgCommons = upMsgCommons
	m.DownMsgCommons = downMsgCommons
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
