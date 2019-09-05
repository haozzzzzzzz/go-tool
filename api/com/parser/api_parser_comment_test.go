package parser

import (
	"fmt"
	"testing"
)

func TestParseApisFromPkgCommentText(t *testing.T) {
	commonParams, apis, err := ParseApisFromPkgCommentText(
		"",
		"",
		"",
		"",
		"",
		`
/**
@api_doc_common_start
{
	"query_data": {
		"_uid": "用户ID|string|required"
	},
	"header_data": {
		"Content-Type": "内容类型|string|required"
	}
}
@api_doc_common_end

@api_doc_start
{
	"http_method": "GET",
	"relative_paths": ["/hello_world"],
	"query_data": {
		"name": "姓名|string|required"
	},
	"post_data": {
		"location": "地址|string|required"
	},
	"resp_data": {
	    "a": "a|int",
	    "b": "b|int",
	    "c": {
			"d": "d|string"
		},
		"__c": "c|object",
		"f": [
			"string"
		],
		"__f": "f|object|required",
		"g": [
			{
				"h": "h|string|required"
			}
		],
		"__g": "g|array|required"
	}
}
@api_doc_end
	`,
	)
	if nil != err {
		t.Error(err)
		return
	}
	for _, api := range apis {
		fmt.Printf("api: %#v\n", api)
	}
	_ = apis

	for _, params := range commonParams {
		fmt.Printf("params: %#v\n", params)
	}
}
