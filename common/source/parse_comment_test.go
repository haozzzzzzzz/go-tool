package source

import (
	"fmt"
	"testing"
)

func TestNewCommentText(t *testing.T) {
	strCommentText1 := `/***Description Tittle
*@ws_doc_up_common
********line 1
* @ws_doc_down_common
* line 2
* @ws_doc_up_body: 1
* line 3
* @ws_doc_down_body: 2
* line 4
* text 5
sdfsdf */
`
	commentText, err := NewCommentText(strCommentText1)
	if err != nil {
		t.Error(err)
		return
	}

	//fmt.Printf(commentText.String())
	fmt.Println(commentText.String())
}

func TestCommentLineTrim(t *testing.T) {
	texts := []string{
		"//** xxx xxxx xxsdf* *sd */",
		"/*sdfsdfoispdf fsdfsidfu**/",
		"* sdfsdf */",
	}
	for _, text := range texts {
		newText, err := commentLineTrim(text)
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println("new:", newText, "old:", text)
	}
}
