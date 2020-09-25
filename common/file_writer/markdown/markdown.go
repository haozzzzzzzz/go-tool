package markdown

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

type WorkBook struct {
	Buf *bytes.Buffer
}

func NewWorkBook() *WorkBook {
	return &WorkBook{
		Buf: bytes.NewBuffer(nil),
	}
}

// headline
func (m *WorkBook) WriteHeadline(level int, text string) {
	strLevel := strings.Repeat("#", level)
	text = fmt.Sprintf("%s %s\n", strLevel, text)
	m.Buf.WriteString(text)
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
		m.Buf.WriteString(text)
	}
}

func (m *WorkBook) WriteTextLn(texts ...string) {
	m.WriteText(texts...)
	m.Buf.WriteString("\n")
}

// table
func (m *WorkBook) WriteTable(table *Table) {
	m.Buf.WriteString(table.String())
}

// string
func (m *WorkBook) String() (str string) {
	str = m.Buf.String()
	return
}

// bytes
func (m *WorkBook) Bytes() []byte {
	return m.Buf.Bytes()
}

// Save 保存markdown
func (m *WorkBook) Save(filename string) (err error) {
	err = ioutil.WriteFile(filename, m.Buf.Bytes(), os.FileMode(0666))
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
