// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ws

import (
	"bytes"
	"time"
)

type Message struct {
	messageType int
	buffer      *bytes.Buffer
	timeout     time.Duration
}

func NewMessage(messageType int, buffer *bytes.Buffer) *Message {
	return &Message{
		messageType: messageType,
		buffer:      buffer,
	}
}

func NewControlMessage(messageType int, buffer *bytes.Buffer, timeout time.Duration) *Message {
	return &Message{
		messageType: messageType,
		buffer:      buffer,
		timeout:     timeout,
	}
}
