package ws

import "github.com/gorilla/websocket"

// !!! 常见状态码释义
// • 1000 CLOSE_NORMAL 连接正常关闭
// • 1001 CLOSE_GOING_AWAY 终端离开 例如：服务器错误，或者浏览器已经离开此页面
// • 1002 CLOSE_PROTOCOL_ERROR 因为协议错误而中断连接
// • 1003 CLOSE_UNSUPPORTED 端点因为受到不能接受的数据类型而中断连接
// • 1004 保留
// • 1005 CLOSE_NO_STATUS 保留, 用于提示应用未收到连接关闭的状态码
// • 1006 CLOSE_ABNORMAL 期望收到状态码时连接非正常关闭 (也就是说, 没有发送关闭帧)
// • 1007 Unsupported Data 收到的数据帧类型不一致而导致连接关闭
// • 1008 Policy Violation 收到不符合约定的数据而断开连接
// • 1009 CLOSE_TOO_LARGE 收到的消息数据太大而关闭连接
// • 1010 Missing Extension 客户端因为服务器未协商扩展而关闭
// • 1011 Internal Error 服务器因为遭遇异常而关闭连接
// • 1012 Service Restart 服务器由于重启而断开连接
// • 1013 Try Again Later 服务器由于临时原因断开连接, 如服务器过载因此断开一部分客户端连接
// • 1015 TLS握手失败关闭连接

func isControl(frameType int) bool {
	return frameType == websocket.CloseMessage || frameType == websocket.PingMessage || frameType == websocket.PongMessage
}
