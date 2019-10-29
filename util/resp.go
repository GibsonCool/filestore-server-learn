package util

import (
	"encoding/json"
	"fmt"
	"log"
)

// RespMsg: http响应数据的通用结构 使用 struct tags 处理json转换小写问题
type RespMsg struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// JsonToBytes: 结构体转换为 json 格式的二进制数组
func (resp *RespMsg) JsonToBytes() []byte {

	bytes, e := json.Marshal(resp)
	if e != nil {
		log.Println(e.Error())
	}
	return bytes
}

// JsonToBytString: 结构体转换成 json 格式的 string
func (resp *RespMsg) JsonToBytString() string {
	return string(resp.JsonToBytes())
}

// NewRespMsg: 生成 response 对象
func NewRespMsg(code int, msg string, data interface{}) *RespMsg {
	return &RespMsg{code, msg, data}
}

// GenSimpleRespStream : 只包含code和message的响应体([]byte)
func GenSimpleRespStream(code int, msg string) []byte {
	return []byte(fmt.Sprintf(`{"code":%d,"msg":"%s"}`, code, msg))
}

// GenSimpleRespString : 只包含code和message的响应体(string)
func GenSimpleRespString(code int, msg string) string {
	return fmt.Sprintf(`{"code":%d,"msg":"%s"}`, code, msg)
}
