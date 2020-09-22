package parse

// tag key类型
type WsTagKeyType string

const WsTagUpCommon WsTagKeyType = "ws_doc_up_common"     // 上行消息公共参数
const WsTagDownCommon WsTagKeyType = "ws_doc_down_common" // 下行消息公共参数
const WsTagCommon WsTagKeyType = "ws_doc_common"          // 上下行消息公共参数
const WsTagUpBody WsTagKeyType = "ws_doc_up_body"         // 上行消息消息体
const WsTagDownBody WsTagKeyType = "ws_doc_down_body"     // 下行消息消息体

type WsTagContent struct {
	TagKey string `json:"tag_key"`
}
