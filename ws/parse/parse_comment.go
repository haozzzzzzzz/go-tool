package parse

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"regexp"
	"strings"
)

type CommentText struct {
	Tags  map[string]string `json:"tags"`
	Lines []string          `json:"lines"` // 除了tag，文本行
}

func (m *CommentText) String() (str string) {
	str += "tags:\n"
	for key, value := range m.Tags {
		str += fmt.Sprintf("%s: %s\n", key, value)
	}

	str += "\nlines:\n"
	for _, line := range m.Lines {
		str += fmt.Sprintf("%s\n", line)
	}
	return
}

func NewCommentText(text string) (commentText *CommentText, err error) {
	commentText = &CommentText{
		Tags:  make(map[string]string),
		Lines: make([]string, 0),
	}

	bufReader := bufio.NewReader(strings.NewReader(text))
	tagReg, err := regexp.Compile(`(?i:\@(\w+)\:?)`)
	if err != nil {
		logrus.Errorf("compile comment text tag regexp failed. error: %s", err)
		return
	}

	for {
		bLine, _, errR := bufReader.ReadLine()
		err = errR
		if err != nil && err != io.EOF {
			logrus.Errorf("read comment text line failed. error: %s", err)
			return
		}

		if err == io.EOF {
			err = nil
			break
		}

		strLine := string(bLine)
		strLine, err = commentLineTrim(strLine)
		if err != nil {
			logrus.Errorf("trim comment text line failed. error: %s", err)
			return
		}

		matched := tagReg.FindAllStringSubmatch(strLine, 1)
		if len(matched) != 1 || len(matched[0]) != 2 {
			// keep space and do not trim
			commentText.Lines = append(commentText.Lines, strLine)
			continue
		}

		matchedResult := matched[0]
		matchedText := matchedResult[0]
		matchedTagKey := matchedResult[1]
		matchedTagKey = strings.TrimSpace(matchedTagKey)
		matchedTagValue := strings.Replace(strLine, matchedText, "", 1)
		matchedTagValue = strings.TrimSpace(matchedTagValue)

		commentText.Tags[matchedTagKey] = matchedTagValue
	}

	return
}

/// trim comment char
func commentLineTrim(oldText string) (newText string, err error) {
	newText = strings.TrimSpace(oldText)
	reg, err := regexp.Compile(`(^[\/\*]+\s?)|(\*+\/$)`)
	if err != nil {
		logrus.Errorf("compile trim target regexp failed error: %s", err)
		return
	}
	newText = reg.ReplaceAllString(newText, "")

	return
}
