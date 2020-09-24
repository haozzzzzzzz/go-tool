package source

import (
	"testing"
)

func TestParseTypes(t *testing.T) {
	dir := "/Users/hao/Documents/Projects/Go/ws_doc"
	parser := NewTypesParser(dir)
	err := parser.ParseTypes(func(params *TypeFilterParams) {

	})
	if err != nil {
		t.Error(err)
		return
	}
}
