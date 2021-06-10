package source

import (
	"errors"
	"fmt"
	"testing"
)

func TestParseTypes(t *testing.T) {
	dir := "/Users/hao/Documents/Projects/Go/ws_doc/test1"
	_, parsedTypes, parsedVals, err := Parse(dir)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(parsedTypes)
	fmt.Println(parsedVals)
}

func TestParseTypes2(t *testing.T) {
	//dir := "/Users/hao/Documents/Projects/Go/ws_doc"
	//parsedTypes, parsedVals, err := Parse(dir)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	//
	//for _, parsedType := range parsedTypes {
	//	fmt.Println(parsedType.TypeName)
	//}
	//
	//fmt.Println(parsedTypes)
	//fmt.Println(parsedVals)

	var err error
	ptrErr := &err

	err = errors.New("err1")
	fmt.Println(ptrErr, *ptrErr)

	err = errors.New("err2")
	fmt.Println(ptrErr, *ptrErr)
}
