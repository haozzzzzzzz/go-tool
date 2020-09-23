features:
- Generating websocket protocol document


### websocket doc tag specification
```go

// @ws_doc_msg_id_type
type MsgIdType uint32

const UpMsgIdEnterRoom MsgIdType = 1001         // enter room msg id
const DownMsgIdEnterRoomResult MsgIdType = 2001 // enter room result msg id

// @ws_doc_up_common
// 上行消息公共参数
type UpCommonParams struct {
	MsgId MsgIdType   `json:"msg_id"`
	Body  interface{} `json:"body"`
}

// @ws_doc_down_common
// 下行消息公共参数
type DownCommonParams struct {
	MsgId MsgIdType   `json:"msg_id"`
	Body  interface{} `json:"body"`
}

// @ws_doc_common
type UpDownCommonParams struct {
    Nonce string `json:"nonce"`
}

// @ws_doc_up_body: UpMsgIdEnterRoom
type UpBodyEnterRoom struct {
	UserType uint8 `json:"user_type"` // user type
}

// @ws_doc_down_body: DownMsgIdEnterRoomResult
type DownBodyEnterRoomResult struct {
	Success bool `json:"success"` // enter room success or not
}

// @ws_doc_body 1,2,3
type UpDownBody1_2_3 struct {
    UserId string `json:"user_id"` // user id
}
```