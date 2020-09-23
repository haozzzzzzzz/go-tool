package parse

import (
	"fmt"
	"testing"
)

func TestWsTagFromCommentText(t *testing.T) {
	strComment := `@api_doc_tag
@ws_doc_up_common
@ws_doc_up_body 1
sdfsdf
sdfsdf
sdfs
df
sdf
sdf
`

	wsTag, err := WsTagFromCommentText(strComment)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(wsTag.String())
}
