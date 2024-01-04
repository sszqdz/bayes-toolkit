package ws

import "github.com/gorilla/websocket"

// !!! Defined Status Codes
// https://www.rfc-editor.org/rfc/rfc6455.html#section-7.4.1
func isControl(frameType int) bool {
	return frameType == websocket.CloseMessage || frameType == websocket.PingMessage || frameType == websocket.PongMessage
}
