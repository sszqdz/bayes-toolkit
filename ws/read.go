// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ws

import (
	"bytes"
	"io"

	"github.com/gorilla/websocket"
)

func (c *Conn[T]) readLoop() {
	code := websocket.CloseNoStatusReceived
	defer func() { c.Close(code) }()
	for {
		messageType, buffer, err := c.readMessage()
		if c.isClosing.Load() {
			return
		}
		if err != nil {
			normal := false
			if e, ok := err.(*websocket.CloseError); ok {
				code = e.Code
				normal = websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) // Close normally
			}
			c.handleMessageErr(c, normal, err)
			return
		}
		c.handleMessage(c, messageType, buffer)
	}
}

func (c *Conn[T]) readMessage() (messageType int, buffer *bytes.Buffer, err error) {
	var r io.Reader
	messageType, r, err = c.Conn.NextReader()
	if err != nil {
		return messageType, nil, err
	}

	buffer = getBuffer()
	_, err = buffer.ReadFrom(r)
	return messageType, buffer, err
}
