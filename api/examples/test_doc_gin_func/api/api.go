package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HeaderData struct {
	H1 string `json:"h_1"`
}

type UriData struct {
	Id string `json:"id"`
}

type QueryData struct {
	Q1 string `json:"q_1"` // haha
}

type PostData struct {
	P1 string `json:"p_1"`
}

type RespDataItem struct {
	R1 string `json:"r_1"`
}

/*
测试接口名字
描述1
@api_doc_http_method: GET
描述1

@api_doc_relative_paths: /test/say_hi/:id, /test/say_hi2/:id
描述1

*/
func SayHiGinHandlerFunc(context *gin.Context) {

	headerData := &HeaderData{}
	uriData := &UriData{}
	queryData := &QueryData{}
	postData := &PostData{}

	_ = queryData
	_ = uriData
	_ = headerData
	_ = postData

	// response data
	type ResponseData struct {
		Msg string `json:"msg"` // 问候信息
	}
	respData := make(map[string]*RespDataItem)

	context.JSON(http.StatusOK, respData)
	return
}

func test(func(context *gin.Context)) {

}

func init() {
	test(func(context *gin.Context) {

	})
}
