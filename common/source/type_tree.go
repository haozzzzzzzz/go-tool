package source

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

func TypeTree(
	iType IType,
	Indent string,
	maxLevel int,
) (strTree string) {
	buf := bytes.NewBuffer(nil)
	indentTexts := ITypeIndentTexts(iType, maxLevel, 0, []IType{})
	for _, indentText := range indentTexts {
		prefix := strings.Repeat(Indent, indentText.Level)
		text := fmt.Sprintf("%s%s", prefix, indentText.Text)
		lenDocLines := len(indentText.DocLines)
		if lenDocLines == 1 {
			text += " // " + strings.Join(indentText.DocLines, " ")
		} else if lenDocLines > 1 {
			docText := "/*\n"
			for _, docLine := range indentText.DocLines {
				docText += docLine + "\n"
			}
			docText += "*/\n"
			text = docText + text
		}

		text += "\n"
		buf.WriteString(text)
	}

	strTree = buf.String()
	return
}

// 缩进文本
type IndentText struct {
	Level    int
	Text     string
	DocLines []string
}

func ITypeIndentTexts(
	iType IType,
	maxLevel int,
	curLevel int,
	path ParseTypePath,
) (texts []*IndentText) {
	// check path
	if path.HasParent(iType) { // 存在循环，直接返回类型
		texts = []*IndentText{
			{
				Level:    curLevel,
				Text:     fmt.Sprintf("<%s>", iType.TypeName()),
				DocLines: []string{},
			},
		}
		return
	}

	nextPath := append(path, iType)
	switch t := iType.(type) {
	case *BasicType:
		texts = BasicTypeIndentTexts(t, maxLevel, curLevel, nextPath)

	case *StructType:
		texts = StructTypeIndentTexts(t, maxLevel, curLevel, nextPath)

	case *MapType:
		texts = MapTypeIndentTexts(t, maxLevel, curLevel, nextPath)

	case *ArrayType:
		texts = ArrayTypeIndentTexts(t, maxLevel, curLevel, nextPath)

	case *InterfaceType:
		texts = InterfaceTypeIndentTexts(t, maxLevel, curLevel, nextPath)

	default:
		texts = make([]*IndentText, 0)
		logrus.Warnf("ITypeIndentTexts unsupported type: %#v", iType)
	}
	return
}

// basic_type
// ignore max level
func BasicTypeIndentTexts(
	basicType *BasicType,
	maxLevel int,
	curLevel int,
	path ParseTypePath,
) (texts []*IndentText) {
	indentText := &IndentText{
		Level:    curLevel,
		Text:     fmt.Sprintf("<%s>", basicType.TypeName()),
		DocLines: []string{},
	}

	if basicType.Description != "" {
		indentText.DocLines = append(indentText.DocLines, basicType.Description)
	}

	texts = []*IndentText{indentText}
	return
}

// field
// 如果出现循环引用，则直接返回，不再继续查找
func FieldIndentTexts(
	field *Field,
	maxLevel int,
	curLevel int,
	path ParseTypePath,
) (texts []*IndentText) {
	texts = make([]*IndentText, 0)
	indentText := &IndentText{
		Level:    curLevel,
		Text:     fmt.Sprintf("%s: ", field.TagJsonOrName()),
		DocLines: []string{},
	}

	texts = append(texts, indentText)
	if curLevel >= maxLevel {
		indentText.Text += field.TypeName
		if field.Description != "" {
			indentText.DocLines = append(indentText.DocLines, field.Description)
		}
		return
	}

	subTexts := ITypeIndentTexts(field.TypeSpec, maxLevel, curLevel, path)
	if len(subTexts) > 0 {
		indentText.Text += subTexts[0].Text
		subTexts = subTexts[1:]
	}

	if field.Description != "" {
		indentText.DocLines = append(indentText.DocLines, field.Description)
	}

	texts = append(texts, subTexts...)
	return
}

func StructTypeIndentTexts(
	structType *StructType,
	maxLevel int,
	curLevel int,
	path ParseTypePath,
) (texts []*IndentText) {
	texts = make([]*IndentText, 0)

	if curLevel >= maxLevel {
		texts = append(texts, &IndentText{
			Level:    curLevel,
			Text:     fmt.Sprintf("%s %s", structType.TypeName(), structType.TypeClass),
			DocLines: []string{},
		})
		return
	}

	texts = append(texts, &IndentText{
		Level:    curLevel,
		Text:     "{",
		DocLines: []string{},
	})

	for _, field := range structType.Fields {
		texts = append(texts, FieldIndentTexts(field, maxLevel, curLevel+1, path)...)
	}

	texts = append(texts, &IndentText{
		Level:    curLevel,
		Text:     "}",
		DocLines: []string{},
	})

	return
}

func MapTypeIndentTexts(
	mapType *MapType,
	maxLevel int,
	curLevel int,
	path ParseTypePath,
) (texts []*IndentText) {
	texts = make([]*IndentText, 0)

	if curLevel >= maxLevel {
		texts = append(texts, &IndentText{
			Level:    curLevel,
			Text:     fmt.Sprintf("%s %s", mapType.TypeName(), mapType.TypeClass),
			DocLines: []string{},
		})
		return
	}

	mapLeftBracket := &IndentText{
		Level: curLevel,
		Text:  fmt.Sprintf("{"),
	}
	texts = append(texts, mapLeftBracket)

	mapInnerBlock := &IndentText{
		Level:    curLevel + 1,
		Text:     fmt.Sprintf("<%s>: ", mapType.KeySpec.TypeName()),
		DocLines: []string{},
	}

	texts = append(texts, mapInnerBlock)
	subTexts := ITypeIndentTexts(mapType.ValueSpec, maxLevel, curLevel+1, path)
	if len(subTexts) > 0 {
		mapInnerBlock.Text += subTexts[0].Text
		subTexts = subTexts[1:]
	}

	texts = append(texts, subTexts...)

	texts = append(texts, &IndentText{
		Level:    curLevel,
		Text:     "}",
		DocLines: []string{},
	})

	return
}

func ArrayTypeIndentTexts(
	arrayType *ArrayType,
	maxLevel int,
	curLevel int,
	path ParseTypePath,
) (texts []*IndentText) {
	texts = make([]*IndentText, 0)
	if curLevel >= maxLevel {
		texts = append(texts, &IndentText{
			Level:    curLevel,
			Text:     fmt.Sprintf("%s %s", arrayType.TypeName(), arrayType.TypeClass),
			DocLines: []string{},
		})
		return
	}

	texts = append(texts, &IndentText{
		Level:    curLevel,
		Text:     "[",
		DocLines: []string{},
	})

	texts = append(texts, ITypeIndentTexts(arrayType.EltSpec, maxLevel, curLevel+1, path)...)

	texts = append(texts, &IndentText{
		Level:    curLevel,
		Text:     "]",
		DocLines: []string{},
	})
	return
}

func InterfaceTypeIndentTexts(
	interfaceType *InterfaceType,
	maxLevel int,
	curLevel int,
	path ParseTypePath,
) (texts []*IndentText) {
	texts = append(texts, &IndentText{
		Level:    curLevel,
		Text:     "<variable>",
		DocLines: []string{},
	})
	return
}

// 循环引用检测
type ParseTypePath []IType

func NewParseTypePath(path []IType) ParseTypePath {
	return path
}

func (m ParseTypePath) HasParent(
	iType IType,
) (hasParent bool) {
	for _, parent := range m {
		if parent == iType {
			hasParent = true
			break
		}
	}
	return
}
