package api

import (
	"github.com/haozzzzzzzz/go-rapid-development/web/ginbuilder"
)

type A struct {
	A1 string `json:"a_1"`
}

type B struct {
	A
	B1 string `json:"b_1"`
}

type C struct {
	A
	C1 string `json:"c_1" binding:"required"`
}

type D struct {
	B
	C
	D1 string `json:"d_1"`
}

var Test ginbuilder.HandleFunc = ginbuilder.HandleFunc{
	HttpMethod: "GET",
	RelativePaths: []string{
		"/test",
	},
	Handle: func(ctx *ginbuilder.Context) (err error) {
		// response data
		type ResponseData struct {
			E *D `json:"e"`
		}
		respData := &ResponseData{}
		ctx.SuccessReturn(respData)
		return
	},
}
