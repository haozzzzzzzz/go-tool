package source

import "testing"

func TestParseTypes(t *testing.T) {
	err := parseTypes("/Users/hao/Documents/Projects/Go/ws_doc")
	if err != nil {
		t.Error(err)
		return
	}
}
