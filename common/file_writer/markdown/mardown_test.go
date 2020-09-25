package markdown

import (
	"testing"
)

func TestTable(t *testing.T) {
	tbl := &Table{
		Headers: []string{"标题1", "标题2", "标题3"},
		Rows: [][]string{
			{
				"data_1_1", "data_1_2", "data_1_3",
			},
			{
				"data_2_1", "data_2_2", "data_2_3",
			},
			{
				"data_3_1", "data_3_2", "data_3_3",
			},
		},
	}
	md := NewWorkBook()
	md.WriteH1("hello")
	md.WriteTable(tbl)
	err := md.Save("./test.md")
	if err != nil {
		t.Error(err)
		return
	}
}
