package parser

import (
	"fmt"
	"testing"
)

func TestMergeApiItemParams(t *testing.T) {
	params := MergeApiItemParams(&ApiItemParams{
		HeaderData: &StructType{
			Fields: []*Field{
				{
					Name: "xxx",
				},
			},
		},
	}, &ApiItemParams{
		HeaderData: &StructType{
			Fields: []*Field{
				{
					Name: "666",
				},
			},
		},
	})
	fmt.Printf("%#v\n", params)
}
