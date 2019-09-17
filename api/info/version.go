package info

import (
	"fmt"
)

var BuildTime = "" // 构建时间，由链接器传入

func Info() string {
	return fmt.Sprintf("build_time: %s", BuildTime)
}
