package precompiler

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"regexp"
	"testing"
)

func TestPrecompileText(t *testing.T) {
	_, newFileText, err := PrecompileText("/Users/hao/Documents/Projects/Github/go-tool/code/example/test_code/src/src.pre.go", nil, map[string]interface{}{
		"USE_IMPORT": 3,
	})
	if nil != err {
		t.Error(err)
		return
	}

	fmt.Printf(newFileText)
}

func TestPrecompileFile(t *testing.T) {
	success, newFilename, err := PrecompileFile("/Users/hao/Documents/Projects/Github/go-tool/code/example/test_code/src/src.pre.go", map[string]interface{}{
		//"USE_IMPORT": 3,
	})
	if nil != err {
		t.Error(err)
		return
	}

	_ = success
	fmt.Println(newFilename)
}

func TestRe(t *testing.T) {
	reg, err := regexp.Compile(`//\s*\+build\s*ignore`)
	if nil != err {
		logrus.Errorf("regexp compile failed. error: %s.", err)
		return
	}

	match := reg.Match([]byte(`// +build   ignore`))
	if match {
		fmt.Println("match")
	} else {
		fmt.Println("not match")
	}
}

func TestPrecompile(t *testing.T) {
	err := Precompile("/Users/hao/Documents/Projects/Github/go-tool/code", map[string]interface{}{})
	if nil != err {
		t.Error(err)
		return
	}
}
