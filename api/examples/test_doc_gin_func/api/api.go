package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
sfaasdfasd
@api_doc_http_method: GET
@api_doc_relative_paths: /test/say_hi
*/
func SayHiGinHandlerFunc(context *gin.Context) {
	// response data
	type ResponseData struct {
		Msg string `json:"msg"` // 问候信息
	}
	respData := &ResponseData{
		Msg: "hi",
	}

	context.JSON(http.StatusOK, respData)
	return
}

func test(func(context *gin.Context)) {

}

func init() {
	test(func(context *gin.Context) {

	})
}
