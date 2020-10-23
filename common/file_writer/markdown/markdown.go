package markdown

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type WorkBook struct {
	TOC            *TOC
	buf            *bytes.Buffer
	anchorCountMap map[string]int
}

func NewWorkBook(
	tocMaxLevel int,
) *WorkBook {
	return &WorkBook{
		buf:            bytes.NewBuffer(nil),
		TOC:            NewTOC(tocMaxLevel),
		anchorCountMap: map[string]int{},
	}
}

// headline
func (m *WorkBook) WriteHeadline(level int, text string) {
	anchor := text
	replacedExpr, err := regexp.Compile(`(?i:[^\s\w\p{Han}])`) // 汉语 chinese
	if err != nil {
		logrus.Errorf("compile replace anchor char reg expr failed. error: %s", err)

	} else {
		anchor = replacedExpr.ReplaceAllString(anchor, "")
		anchor = strings.ReplaceAll(anchor, " ", "-")

	}

	counter := m.anchorCountMap[anchor]
	counter++
	m.anchorCountMap[anchor] = counter
	if counter > 1 {
		anchor = fmt.Sprintf("%s-%d", anchor, counter)
	}

	anchor = strings.ToLower(anchor)
	m.TOC.Push(&TOCNode{
		Level:  level,
		Text:   text,
		Anchor: "#" + anchor,
	})

	strLevel := strings.Repeat("#", level)
	text = fmt.Sprintf("%s %s\n", strLevel, text)
	m.buf.WriteString(text)
}

func (m *WorkBook) WriteH1(text string) {
	m.WriteHeadline(1, text)
}

func (m *WorkBook) WriteH2(text string) {
	m.WriteHeadline(2, text)
}

func (m *WorkBook) WriteH3(text string) {
	m.WriteHeadline(3, text)
}

func (m *WorkBook) WriteH4(text string) {
	m.WriteHeadline(4, text)
}

// text
func (m *WorkBook) WriteText(texts ...string) {
	for _, text := range texts {
		m.buf.WriteString(text)
	}
}

func (m *WorkBook) WriteTextLn(texts ...string) {
	m.WriteText(texts...)
	m.buf.WriteString("\n")
}

// table
func (m *WorkBook) WriteTable(table *Table) {
	m.buf.WriteString(table.String())
}

// string
func (m *WorkBook) String() (str string) {
	toc := m.TOC.String()
	if toc != "" {
		str += strings.Repeat("#", m.TOC.RootLevel) + " 目录\n"
		str += toc + "\n"
	}

	str += m.buf.String()
	return
}

// bytes
func (m *WorkBook) Bytes() []byte {
	return m.buf.Bytes()
}

// Save 保存markdown
func (m *WorkBook) Save(filename string) (err error) {
	err = ioutil.WriteFile(filename, []byte(m.String()), os.FileMode(0666))
	if err != nil {
		logrus.Errorf("write markdown workbook failed. error: %s", err)
		return
	}
	return
}

type Table struct {
	Headers []string
	Rows    [][]string
}

func (m *Table) String() (str string) {
	strHead := strings.Join(m.Headers, " | ")
	strHead = fmt.Sprintf("| %s |\n", strHead)
	strSplit := strings.Repeat(" ------ |", len(m.Headers))
	strSplit = fmt.Sprintf("|%s\n", strSplit)

	strBody := ""
	for _, row := range m.Rows {
		strRow := strings.Join(row, " | ")
		strRow = fmt.Sprintf("| %s |\n", strRow)
		strBody += strRow
	}

	str = strHead + strSplit + strBody
	return
}

// 文档目录
type TOC struct {
	MaxLevel  int
	nodes     []*TOCNode
	RootLevel int // 顶层的level
}

type TOCNode struct {
	Level  int // 大于0
	Text   string
	Anchor string
}

func NewTOC(maxLevel int) *TOC {
	return &TOC{
		MaxLevel: maxLevel,
		nodes:    make([]*TOCNode, 0),
	}
}

func (m *TOC) Push(node *TOCNode) {
	if node.Level < 1 {
		return
	}

	m.nodes = append(m.nodes, node)
	if node.Level < m.RootLevel || m.RootLevel == 0 {
		m.RootLevel = node.Level
	}
}

func (m *TOC) String() (str string) {
	for _, node := range m.nodes {
		indentNum := node.Level - m.RootLevel
		str += fmt.Sprintf("%s- [%s](%s)\n", strings.Repeat("\t", indentNum), node.Text, node.Anchor)
	}
	return
}
