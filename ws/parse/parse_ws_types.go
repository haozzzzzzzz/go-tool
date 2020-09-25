package parse

import (
	"github.com/haozzzzzzzz/go-tool/common/source"
	"github.com/sirupsen/logrus"
	"strings"
)

type WsTypesParser struct {
	parser *source.TypesParser

	wsTypes *WsTypes
}

func NewWsTypesParser(
	rootDir string,
) *WsTypesParser {
	return &WsTypesParser{
		parser:  source.NewTypesParser(rootDir),
		wsTypes: NewWsTypes(),
	}
}

// 解析含有ws标签的类型
func (m *WsTypesParser) ParseWsTypes() (err error) {
	parsedTypes, parsedVals, err := m.parser.Parse()
	if err != nil {
		logrus.Errorf("parse source types failed. error: %s", err)
		return
	}

	for _, parsedType := range parsedTypes {
		err = m.typeFilter(parsedType)
		if err != nil {
			logrus.Errorf("ParseWsType typeFilter failed. error: %s", err)
			return
		}
	}

	if m.wsTypes.MsgIdType != nil {
		// find msg id value
		for _, parsedVal := range parsedVals {
			err = m.msgIdFilter(parsedVal)
			if err != nil {
				logrus.Errorf("ParseWsType msgIdFilter failed. error: %s", err)
				return
			}
		}

		// msg_id remap body
		err = m.msgIdRemapBody()
		if err != nil {
			logrus.Errorf("ParseWsType msg id remap body failed. error: %s", err)
			return
		}
	}

	return
}

func (m *WsTypesParser) typeFilter(parsedType *source.ParsedType) (err error) {
	wsTag, err := WsTagFromCommentText(parsedType.Doc)
	if err != nil {
		logrus.Errorf("parse ws tag failed. error: %s", err)
		return
	}

	if !wsTag.Valid { // 不包含ws的标签信息
		return
	}

	typeIdent := source.ParseTypeName(parsedType.TypesInfo, parsedType.TypeName)

	if wsTag.IsMsgIdType {
		m.wsTypes.MsgIdType = typeIdent
	}

	upMsgs := m.wsTypes.UpMsgs
	downMsg := m.wsTypes.DownMsgs

	if wsTag.HasUpCommon {
		upMsgs.Commons = append(upMsgs.Commons, &MsgCommon{
			TypeIdent: typeIdent,
			DocLines:  wsTag.Doc.Lines,
			Comment:   parsedType.Comment,
		})
	}

	if wsTag.HasDownCommon {
		downMsg.Commons = append(upMsgs.Commons, &MsgCommon{
			TypeIdent: typeIdent,
			DocLines:  wsTag.Doc.Lines,
			Comment:   parsedType.Comment,
		})
	}

	for _, body := range wsTag.UpBodys {
		for _, msgId := range body.MsgIds {
			upMsgs.StrMsgIdMapBody[msgId] = &MsgBody{
				StrMsgId:  msgId,
				MsgId:     nil,
				TypeIdent: typeIdent,
				DocLines:  wsTag.Doc.Lines,
				Comment:   parsedType.Comment,
			}
		}
	}

	for _, body := range wsTag.DownBodys {
		for _, msgId := range body.MsgIds {
			downMsg.StrMsgIdMapBody[msgId] = &MsgBody{
				StrMsgId:  msgId,
				MsgId:     nil,
				TypeIdent: typeIdent,
				DocLines:  wsTag.Doc.Lines,
				Comment:   parsedType.Comment,
			}
		}
	}

	return
}

// parse msg id
func (m *WsTypesParser) msgIdFilter(parsedVal *source.ParsedVal) (err error) {
	if parsedVal.Type != m.wsTypes.MsgIdType.TypeName.Type() {
		return
	}

	if parsedVal.Value == "" {
		logrus.Warnf("msg id should has specified value. msg_id: %s", parsedVal.Name)
		return
	}

	m.wsTypes.MsgIds = append(m.wsTypes.MsgIds, parsedVal)
	return
}

// msg_id remap body
func (m *WsTypesParser) msgIdRemapBody() (err error) {
	for _, msgId := range m.wsTypes.MsgIds {
		strMsgIds := []string{msgId.Name, msgId.Value}
		for _, strMsgId := range strMsgIds {
			upBody, ok := m.wsTypes.UpMsgs.StrMsgIdMapBody[strMsgId]
			if ok {
				upBody.MsgId = msgId
			}

			downBody, ok := m.wsTypes.DownMsgs.StrMsgIdMapBody[strMsgId]
			if ok {
				downBody.MsgId = msgId
			}
		}
	}
	return
}

func (m *WsTypesParser) WsTypes() *WsTypes {
	return m.wsTypes
}

// ws types
type UpDownMsgs struct {
	Commons         []*MsgCommon
	StrMsgIdMapBody map[string]*MsgBody
}

type MsgCommon struct {
	TypeIdent *source.TypeIdent
	DocLines  []string
	Comment   string
}

type MsgBody struct {
	StrMsgId  string
	MsgId     *source.ParsedVal
	TypeIdent *source.TypeIdent
	DocLines  []string
	Comment   string
}

func NewUpDownMsgs() *UpDownMsgs {
	return &UpDownMsgs{
		Commons:         make([]*MsgCommon, 0),
		StrMsgIdMapBody: map[string]*MsgBody{},
	}
}

type WsTypes struct {
	MsgIdType *source.TypeIdent
	MsgIds    []*source.ParsedVal

	UpMsgs   *UpDownMsgs
	DownMsgs *UpDownMsgs
}

func NewWsTypes() *WsTypes {
	return &WsTypes{
		MsgIds:   []*source.ParsedVal{},
		UpMsgs:   NewUpDownMsgs(),
		DownMsgs: NewUpDownMsgs(),
	}
}

func (m *WsTypes) Output() (output *WsTypesOutput) {
	output = NewWsTypesOutput()

	for _, msgId := range m.MsgIds {
		oMsgId := &WsMsgIdOutput{
			Name:  msgId.Name,
			Value: msgId.Value,
			Title: "",
			Doc:   "",
			IType: m.MsgIdType.IType,
		}

		docs := make([]string, 0)
		if msgId.Doc != "" {
			docs = append(docs, msgId.Doc)
		}

		if msgId.Comment != "" {
			docs = append(docs, msgId.Comment)
		}

		if len(docs) > 0 {
			oMsgId.Title = docs[0]
			docs = docs[1:]
			oMsgId.Doc = strings.Join(docs, "\n")
		}

		output.MsgIds = append(output.MsgIds, oMsgId)
	}

	for _, upCommon := range m.UpMsgs.Commons {
		commonOut := &WsMsgCommonOutput{
			Doc:   "",
			Title: "",
			IType: upCommon.TypeIdent.IType,
		}

		docs := upCommon.DocLines
		if upCommon.Comment != "" {
			docs = append(docs, upCommon.Comment)
		}
		if len(docs) > 0 {
			commonOut.Title = docs[0]
			docs = docs[1:]
			commonOut.Doc = strings.Join(docs, "\n")
		}

		output.UpMsgCommons = append(output.UpMsgCommons, commonOut)
	}

	for _, downCommon := range m.DownMsgs.Commons {
		commonOut := &WsMsgCommonOutput{
			Doc:   "",
			Title: "",
			IType: downCommon.TypeIdent.IType,
		}

		docs := downCommon.DocLines
		if downCommon.Comment != "" {
			docs = append(docs, downCommon.Comment)
		}

		if len(docs) > 0 {
			commonOut.Title = docs[0]
			docs = docs[1:]
			commonOut.Doc = strings.Join(docs, "\n")
		}

		output.DownMsgCommons = append(output.DownMsgCommons, commonOut)
	}

	for _, upBody := range m.UpMsgs.StrMsgIdMapBody {
		if upBody.MsgId == nil {
			logrus.Warnf("can not find msg id declaration for msg body. msg_body: %s, msg_id: %s", upBody.TypeIdent.IType.TypeName(), upBody.StrMsgId)
			continue
		}

		oBody := &WsMsgBodyOutput{
			MsgId: upBody.MsgId.Value,
			IType: upBody.TypeIdent.IType,
			Doc:   "",
			Title: "",
		}

		docs := upBody.DocLines
		if upBody.Comment != "" {
			docs = append(docs, upBody.Comment)
		}

		if len(docs) > 0 {
			oBody.Title = docs[0]
			docs = docs[1:]
			oBody.Doc = strings.Join(docs, "\n")
		}

		output.UpMsgBodys = append(output.UpMsgBodys, oBody)
	}

	for _, downBody := range m.DownMsgs.StrMsgIdMapBody {
		if downBody.MsgId == nil {
			logrus.Warnf("can not find msg id declaration for msg body. msg_body: %s, msg_id: %s", downBody.TypeIdent.IType.TypeName(), downBody.StrMsgId)
			continue
		}

		oBody := &WsMsgBodyOutput{
			MsgId: downBody.MsgId.Value,
			IType: downBody.TypeIdent.IType,
			Doc:   "",
			Title: "",
		}

		docs := downBody.DocLines
		if downBody.Comment != "" {
			docs = append(docs, downBody.Comment)
		}
		if len(docs) > 0 {
			oBody.Title = docs[0]
			docs = docs[1:]
			oBody.Doc = strings.Join(docs, "\n")
		}

		output.DownMsgBodys = append(output.DownMsgBodys, oBody)
	}

	output.Sort()
	return
}
