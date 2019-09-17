package api

import (
	"github.com/gin-gonic/gin"
)

// 注意：BindRouters函数体内不能自定义添加任何声明，由api compile命令生成api绑定声明
func BindRouters(engine *gin.Engine) (err error) {
	engine.Handle("GET", "/test/say_hi/:id", SayHiGinHandlerFunc)
	engine.Handle("GET", "/test/say_hi2/:id", SayHiGinHandlerFunc)
	return
}
