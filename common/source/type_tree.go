package source

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

func TypeTree(iType IType, Indent string, maxLevel int) (strTree string) {
	buf := bytes.NewBuffer(nil)
	indentTexts := ITypeIndentTexts(iType, maxLevel, 0)
	for _, indentText := range indentTexts {
		prefix := strings.Repeat(Indent, indentText.Level)
		text := fmt.Sprintf("%s %s\n", prefix, indentText.Text)
		buf.WriteString(text)
	}

	strTree = buf.String()
	return
}

// 缩进文本
type IndentText struct {
	Level int
	Text  string
}

func ITypeIndentTexts(iType IType, maxLevel int, curLevel int) (texts []*IndentText) {
	switch t := iType.(type) {
	case *BasicType:
		texts = BasicTypeIndentTexts(t, maxLevel, curLevel)
	case *StructType:
		texts = StructTypeIndentTexts(t, maxLevel, curLevel)
	case *MapType:
		texts = MapTypeIndentTexts(t, maxLevel, curLevel)
	case *ArrayType:
		texts = ArrayTypeIndentTexts(t, maxLevel, curLevel)
	case *InterfaceType:
		texts = InterfaceTypeIndentTexts(t, maxLevel, curLevel)
	default:
		logrus.Warnf("ITypeIndentTexts unsupported type: %#v", iType)
	}
	return
}

// basic_type
// ignore max level
func BasicTypeIndentTexts(basicType *BasicType, maxLevel int, curLevel int) (texts []*IndentText) {
	texts = []*IndentText{
		{
			Level: curLevel,
			Text:  fmt.Sprintf("<%s>", basicType.TypeName()),
		},
	}
	return
}

// field
func FieldIndentTexts(field *Field, maxLevel int, curLevel int) (texts []*IndentText) {
	texts = make([]*IndentText, 0)
	indentText := &IndentText{
		Level: curLevel,
		Text:  fmt.Sprintf("%s: ", field.TagJsonOrName()),
	}

	texts = append(texts, indentText)
	if curLevel >= maxLevel {
		indentText.Text += field.TypeName
		if field.Description != "" {
			indentText.Text += field.Description
		}
		return
	}

	//subTexts := field.TypeSpec.IndentTexts(maxLevel, curLevel)
	subTexts := ITypeIndentTexts(field.TypeSpec, maxLevel, curLevel)
	if len(subTexts) > 0 {
		indentText.Text += subTexts[0].Text
		subTexts = subTexts[1:]
	}

	if field.Description != "" {
		indentText.Text += " // " + field.Description
	}

	texts = append(texts, subTexts...)
	return
}

func StructTypeIndentTexts(structType *StructType, maxLevel int, curLevel int) (texts []*IndentText) {
	texts = make([]*IndentText, 0)

	if curLevel >= maxLevel {
		texts = append(texts, &IndentText{
			Level: curLevel,
			Text:  fmt.Sprintf("%s %s", structType.TypeName(), structType.TypeClass),
		})
		return
	}

	texts = append(texts, &IndentText{
		Level: curLevel,
		Text:  "{",
	})

	for _, field := range structType.Fields {
		texts = append(texts, FieldIndentTexts(field, maxLevel, curLevel+1)...)
	}

	texts = append(texts, &IndentText{
		Level: curLevel,
		Text:  "}",
	})

	return
}

func MapTypeIndentTexts(mapType *MapType, maxLevel int, curLevel int) (texts []*IndentText) {
	texts = make([]*IndentText, 0)

	if curLevel >= maxLevel {
		texts = append(texts, &IndentText{
			Level: curLevel,
			Text:  fmt.Sprintf("%s %s", mapType.TypeName(), mapType.TypeClass),
		})
		return
	}

	mapLeftBracket := &IndentText{
		Level: curLevel,
		Text:  fmt.Sprintf("{"),
	}
	texts = append(texts, mapLeftBracket)

	mapInnerBlock := &IndentText{
		Level: curLevel + 1,
		Text:  fmt.Sprintf("<%s>: ", mapType.KeySpec.TypeName()),
	}
	texts = append(texts, mapInnerBlock)
	//subTexts := mapType.ValueSpec.IndentTexts(maxLevel, curLevel+1)
	subTexts := ITypeIndentTexts(mapType.ValueSpec, maxLevel, curLevel+1)
	if len(subTexts) > 0 {
		mapInnerBlock.Text += subTexts[0].Text
		subTexts = subTexts[1:]
	}

	texts = append(texts, subTexts...)

	texts = append(texts, &IndentText{
		Level: curLevel,
		Text:  "}",
	})
	return
}

func ArrayTypeIndentTexts(arrayType *ArrayType, maxLevel int, curLevel int) (texts []*IndentText) {
	texts = make([]*IndentText, 0)
	if curLevel >= maxLevel {
		texts = append(texts, &IndentText{
			Level: curLevel,
			Text:  fmt.Sprintf("%s %s", arrayType.TypeName(), arrayType.TypeClass),
		})
		return
	}

	texts = append(texts, &IndentText{
		Level: curLevel,
		Text:  "[",
	})

	texts = append(texts, ITypeIndentTexts(arrayType.EltSpec, maxLevel, curLevel+1)...)

	texts = append(texts, &IndentText{
		Level: curLevel,
		Text:  "]",
	})
	return
}

func InterfaceTypeIndentTexts(interfaceType *InterfaceType, maxLevel int, curLevel int) (texts []*IndentText) {
	texts = append(texts, &IndentText{
		Level: curLevel,
		Text:  "<variable>",
	})
	return
}
