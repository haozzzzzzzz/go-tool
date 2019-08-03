/**
go source precompiler package
*/
package precompiler

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gosexy/to"
	"github.com/haozzzzzzzz/go-rapid-development/utils/uerrors"
	"github.com/haozzzzzzzz/go-tool/api/com/project"
	"github.com/haozzzzzzzz/go-tool/lib/gofmt"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

func PrecompileText(
	filename string,
	src interface{},
	params map[string]interface{},
) (
	compiled bool,
	newFileText string,
	err error,
) {
	f, err := parser.ParseFile(token.NewFileSet(), filename, src, parser.ParseComments)
	if nil != err {
		logrus.Errorf("parser parse file failed. error: %s.", err)
		return
	}

	// pos是从1开始的
	type Block struct {
		Text       string
		StartIndex int
		EndIndex   int // 不包含EndIndex所在的Index
	}

	preStatements := make([]*Block, 0)

	for _, commentGroup := range f.Comments {
		for _, comment := range commentGroup.List {
			strStatement := strings.Replace(comment.Text, "//", "", 1)
			strStatement = strings.TrimSpace(strStatement)
			if !strings.HasPrefix(strStatement, "+pre") {
				continue
			}
			strStatement = strings.Replace(strStatement, "+pre", "", 1)
			strStatement = strings.TrimSpace(strStatement)

			statement := &Block{
				Text:       strStatement,
				StartIndex: int(comment.Pos()) - 1,
				EndIndex:   int(comment.End()),
			}

			preStatements = append(preStatements, statement)
		}
	}

	if len(preStatements) == 0 { // no +pre tag
		return
	}

	bFileText, err := ioutil.ReadFile(filename)
	if nil != err {
		logrus.Errorf("read file failed. error: %s.", err)
		return
	}
	lenOfFile := len(bFileText)
	_ = lenOfFile

	goBlocks := make([]*Block, 0)
	luaBlocks := make([]*Block, 0)

	var lastStatement *Block
	var blocksStart, blocksEnd int
	_ = blocksEnd
	for _, statement := range preStatements {
		if lastStatement == nil {
			blocksStart = statement.StartIndex
			lastStatement = statement

		} else {
			blockText := string(bFileText[lastStatement.EndIndex:statement.StartIndex])
			block := &Block{
				Text:       blockText,
				StartIndex: lastStatement.EndIndex,
				EndIndex:   statement.StartIndex,
			}

			idxGoBlocks := len(goBlocks)
			goBlocks = append(goBlocks, block)

			luaBlock := &Block{
				StartIndex: block.StartIndex,
				EndIndex:   block.EndIndex,
			}
			luaBlocks = append(luaBlocks, luaBlock)

			luaBlock.Text = fmt.Sprintf("use_go_block(%d)\n", idxGoBlocks)

			lastStatement = statement
		}

		blocksEnd = lastStatement.EndIndex
		luaBlocks = append(luaBlocks, statement)
	}

	var luaText string
	for _, block := range luaBlocks {
		luaText += block.Text + "\n"
	}

	Lua := lua.NewState()
	defer Lua.Close()

	useIndexes := make([]int, 0)
	Lua.SetGlobal("use_go_block", Lua.NewFunction(func(state *lua.LState) int {
		lv := state.ToInt(1) // get first argument
		useIndexes = append(useIndexes, lv)
		return 0
	}))

	for key, param := range params {
		var val lua.LValue
		kind := reflect.TypeOf(param).Kind()
		switch {
		case kind == reflect.Bool:
			val = lua.LBool(param.(bool))
		case (kind >= reflect.Int && kind <= reflect.Uint64) || (kind >= reflect.Float32 && kind <= reflect.Float64):
			val = lua.LNumber(to.Int64(param))
		case kind == reflect.String:
			val = lua.LString(to.String(param))
		default:
			err = uerrors.Newf("unsupported go type to lua type. %s", kind)
		}

		Lua.SetGlobal(key, val)
	}

	//fmt.Printf(luaText)

	err = Lua.DoString(luaText)
	if nil != err {
		logrus.Errorf("lua do string failed. error: %s.", err)
		return
	}

	useGoMap := make(map[int]bool)
	for _, useIdx := range useIndexes {
		useGoMap[useIdx] = true
	}

	bNewFileText := make([]byte, 0)

	if blocksStart > 0 {
		bNewFileText = bFileText[:blocksStart]
	}

	for idx, goBlock := range goBlocks {
		if useGoMap[idx] { // 包含
			bNewFileText = append(bNewFileText, bFileText[goBlock.StartIndex:goBlock.EndIndex]...)
		}
	}

	if len(bNewFileText) < blocksEnd {
		bNewFileText = append(bNewFileText, bFileText[blocksEnd:]...)
	}

	newFileText = string(bNewFileText)
	compiled = true
	return
}

const PrecompileFileSuffix = ".pre.go"

func PrecompileFile(filename string, params map[string]interface{}) (success bool, compiledFilename string, err error) {
	filename, err = filepath.Abs(filename)
	if nil != err {
		logrus.Errorf("get abs filepath failed. filename: %s, error: %s.", filename, err)
		return
	}

	baseName := filepath.Base(filename)
	if !strings.HasSuffix(baseName, PrecompileFileSuffix) {
		return
	}

	fileDir := filepath.Dir(filename)
	newBaseName := strings.Replace(baseName, PrecompileFileSuffix, ".go", 1)
	compiledFilename = fmt.Sprintf("%s/%s", fileDir, newBaseName)

	compiled, compiledFileText, err := PrecompileText(filename, nil, params)
	if nil != err {
		logrus.Errorf("precompile file text failed. filename: %s, error: %s.", filename, err)
		return
	}

	if compiled == false {
		return
	}

	logrus.Printf("Precompiled %s \n", filename)

	// add build ignore to original file
	bPreFileText, err := ioutil.ReadFile(filename)
	if nil != err {
		logrus.Errorf("read need precompile file text failed. filename: %s, error: %s.", filename, err)
		return
	}

	// check pre file whether has build ignore or not
	preFileText := string(bPreFileText)
	preFileReader := bufio.NewReader(bytes.NewBuffer(bPreFileText))
	bLine, isPrefix, err := preFileReader.ReadLine()
	if nil != err {
		logrus.Errorf("read pre file first line failed. error: %s.", err)
		return
	}

	if isPrefix {
		err = uerrors.Newf("too long for go package fist line")
		return
	}

	reg, err := regexp.Compile(`//\s*\+build\s*ignore`)
	if nil != err {
		logrus.Errorf("regexp compile failed. error: %s.", err)
		return
	}

	if !reg.Match(bLine) {
		preFileText = fmt.Sprintf("// +build ignore \n\n%s", preFileText)
		preFileText, err = gofmt.StrGoFmt(preFileText)
		if nil != err {
			logrus.Errorf("go fmt pre file text failed. filename: %s, error: %s.", filename, err)
			return
		}

		// save pre file text
		err = ioutil.WriteFile(filename, []byte(preFileText), project.ProjectFileMode)
		if nil != err {
			logrus.Errorf("write file failed. filename: %s, error: %s.", filename, err)
			return
		}
	}

	// compiled text
	bCompiledText := []byte(compiledFileText)
	compiledReader := bufio.NewReader(bytes.NewBuffer(bCompiledText))
	bLine, isPrefix, err = compiledReader.ReadLine()
	if nil != err {
		logrus.Errorf("read first line of compiled file failed. error: %s.", err)
		return
	}

	if isPrefix {
		err = uerrors.Newf("too long for go package fist line")
		return
	}

	// rm potential build ignore tag from pre file text
	if reg.Match(bLine) {
		bCompiledText = bCompiledText[len(bLine):]
	}

	// save compiled file text
	compiledFileText = string(bCompiledText)
	compiledFileText, err = gofmt.StrGoFmt(compiledFileText)
	if nil != err {
		logrus.Errorf("fmt go src failed. error: %s.", err)
		return
	}

	err = ioutil.WriteFile(compiledFilename, []byte(compiledFileText), project.ProjectFileMode)
	if nil != err {
		logrus.Errorf("write file failed. filename: %s, error: %s.", compiledFilename, err)
		return
	}

	success = true

	return
}

func Precompile(path string, params map[string]interface{}) (err error) {
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) (wlkErr error) {
		if info.IsDir() {
			return
		}

		if !strings.HasSuffix(info.Name(), PrecompileFileSuffix) {
			return
		}

		_, _, err = PrecompileFile(path, params)
		if nil != err {
			logrus.Errorf("precompile file failed. path: %s, params: %#v, error: %s.", path, params, err)
			return
		}

		return
	})
	if nil != err {
		logrus.Errorf("walk file path failed. path: %s, error: %s.", path, err)
		return
	}

	return
}
