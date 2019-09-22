package parser

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSaveApisSwaggerSpec(t *testing.T) {
	swgSpc := NewSwaggerSpec()
	swgSpc.Apis([]*ApiItem{
		{
			ApiItemParams: ApiItemParams{
				HeaderData: &StructType{
					Fields: []*Field{
						{
							Name:     "xx",
							TypeName: "string",
							Tags: map[string]string{
								"json":    "content-type",
								"header":  "content-type",
								"binding": "required",
							},
						},
					},
				},
				UriData: &StructType{
					Fields: []*Field{
						{
							Name:     "tt",
							TypeName: "int64",
							Tags: map[string]string{
								"json":    "book_id",
								"binding": "required",
							},
						},
					},
				},
			},
			Summary: "书本信息接口",
			//PackageName:    "pack",
			ApiHandlerFunc: "func",
			HttpMethod:     "GET",
			RelativePaths: []string{
				"/api/book/:book_id",
				"/api/book",
			},
		},
	})
	swgSpc.Info(
		"Book shop",
		"book shop api for testing tools",
		"1",
		"haozzzzzzzz",
	)
	err := swgSpc.ParseApis()
	if nil != err {
		t.Error(err)
		return
	}

	out, err := swgSpc.Output()
	if nil != err {
		t.Error(err)
		return
	}

	fmt.Println(string(out))
}

func TestSaveApisSwaggerSpec2(t *testing.T) {
	_, a, err := ParseApis(
		"/data/apps/go/video_buddy_share/manage/api/share",
		true,
		false,
		true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(a[0])

	swgSpc := NewSwaggerSpec()
	swgSpc.Apis(a)
	swgSpc.Info(
		"Book shop",
		"book shop api for testing tools",
		"1",
		"haozzzzzzzz",
	)
	err = swgSpc.ParseApis()
	if nil != err {
		t.Error(err)
		return
	}

	out, err := json.Marshal(swgSpc.Swagger) //goland输出控制台->右键->Show as JSON
	if nil != err {
		t.Error(err)
		return
	}

	fmt.Println(string(out))
}
