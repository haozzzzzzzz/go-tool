package doc

import (
	"fmt"
	"github.com/haozzzzzzzz/go-tool/common/file_writer/markdown"
	"github.com/haozzzzzzzz/go-tool/common/source"
	"github.com/haozzzzzzzz/go-tool/ws/parse"
	"github.com/sirupsen/logrus"
)

func init() {
	BindFileFormatWriter([]string{"md"}, WriteDocMd)
}

const TOCMaxLevel = 4

func WriteDocMd(
	wsTypes *parse.WsTypesOutput,
	format string, // 格式
	filepath string,
) (err error) {
	mdWriter := NewWsTypesMdWriter(wsTypes)
	mdWriter.Parse()
	err = mdWriter.Save(filepath)
	if err != nil {
		logrus.Errorf("save ws types markdown failed. error: %s", err)
		return
	}
	return
}

type WsTypesMdWriter struct {
	wsTypes  *parse.WsTypesOutput
	mdWriter *markdown.WorkBook
}

func NewWsTypesMdWriter(wsTypes *parse.WsTypesOutput) *WsTypesMdWriter {
	return &WsTypesMdWriter{
		wsTypes:  wsTypes,
		mdWriter: markdown.NewWorkBook(TOCMaxLevel),
	}
}

func (m *WsTypesMdWriter) WriteSectionSeperate() {
	m.mdWriter.WriteTextLn("\n\n\n\n\n")
}

func (m *WsTypesMdWriter) Parse() {
	m.WriteMsgIds("上行消息ID列表", m.wsTypes.UpMsgIdValues, m.wsTypes.UpMsgIdMap)
	m.WriteSectionSeperate()
	m.WriteMsgIds("下行消息ID列表", m.wsTypes.DownMsgIdValues, m.wsTypes.DownMsgIdMap)
	m.WriteSectionSeperate()
	m.WriteUpMsg()
	m.WriteSectionSeperate()
	m.WriteDownMsg()
}

func (m *WsTypesMdWriter) Bytes() []byte {
	return m.mdWriter.Bytes()
}

func (m *WsTypesMdWriter) Save(filename string) (err error) {
	return m.mdWriter.Save(filename)
}

func (m *WsTypesMdWriter) WriteUpMsg() {
	m.WriteCommon("上行消息公参", m.wsTypes.UpMsgCommons)
	m.WriteSectionSeperate()
	m.WriteBody("上行消息负载", m.wsTypes.UpMsgBodys)
}

func (m *WsTypesMdWriter) WriteDownMsg() {
	m.WriteCommon("下行消息公参", m.wsTypes.DownMsgCommons)
	m.WriteSectionSeperate()
	m.WriteBody("下行消息负载", m.wsTypes.DownMsgBodys)
}

func (m *WsTypesMdWriter) WriteMsgIds(title string, msgIdValues []string, msgIdMap map[string]*parse.WsMsgIdOutput) {
	if len(msgIdValues) == 0 {
		return
	}

	m.mdWriter.WriteH3(title)
	msgIdTable := &markdown.Table{
		Headers: []string{
			"value", "key", "type", "title", "description",
		},
		Rows: [][]string{},
	}

	for _, msgIdValue := range msgIdValues {
		msgId, ok := msgIdMap[msgIdValue]
		if !ok {
			logrus.Warnf("msg id ident not found. value: %s", msgIdValue)
			continue
		}

		row := []string{
			msgId.Value,
			msgId.Name,
			msgId.IType.TypeName(),
			msgId.Title,
			msgId.Doc,
		}
		msgIdTable.Rows = append(msgIdTable.Rows, row)
	}

	m.mdWriter.WriteTable(msgIdTable)
}

func (m *WsTypesMdWriter) WriteCommon(title string, commonOutputs []*parse.WsMsgCommonOutput) {
	if len(commonOutputs) == 0 {
		return
	}

	m.mdWriter.WriteH3(title)
	for _, commonOutput := range commonOutputs {
		m.mdWriter.WriteH4(fmt.Sprintf("%s %s", commonOutput.IType.TypeName(), commonOutput.Title))
		m.mdWriter.WriteTextLn(commonOutput.Doc)
		m.writeType(commonOutput.IType)
	}
}

func (m *WsTypesMdWriter) WriteBody(title string, bodyOutputs []*parse.WsMsgBodyOutput) {
	if len(bodyOutputs) == 0 {
		return
	}

	m.mdWriter.WriteH3(title)
	for _, bodyOutput := range bodyOutputs {
		m.mdWriter.WriteH4(fmt.Sprintf("[%s] %s %s", bodyOutput.MsgId, bodyOutput.IType.TypeName(), bodyOutput.Title))
		m.mdWriter.WriteTextLn(bodyOutput.Doc)
		m.writeType(bodyOutput.IType)
	}
}

func (m *WsTypesMdWriter) writeType(iType source.IType) {
	strTree := source.TypeTree(iType, "\t", 10)
	if strTree == "" {
		return
	}

	m.mdWriter.WriteText(fmt.Sprintf("```\n%s```\n", strTree))
}
