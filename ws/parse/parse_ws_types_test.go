package parse

import (
	"fmt"
	"testing"
)

func TestWsTypesParser_ParseWsTypes(t *testing.T) {
	dir := "/Users/hao/Documents/Projects/Go/ws_doc"
	wsParser := NewWsTypesParser(dir)
	err := wsParser.ParseWsTypes()
	if err != nil {
		t.Error(err)
		return
	}

	bObj, err := wsParser.WsTypes().Output().Json()
	if err != nil {
		t.Error(err)
		return
	}
	_ = bObj

	fmt.Println(string(bObj))
}
