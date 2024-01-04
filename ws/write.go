package ws

import "github.com/gorilla/websocket"

func (c *Conn[T]) writeLoop() {
	defer close(c.writeEndChan)
	for value := range c.writeChan {
		if c.isClosed.Load() {
			return
		}
		var err error
		switch message := value.(type) {
		case *Message:
			err = c.writeMessage(message)
			putBuffer(message.buffer)
		case *websocket.PreparedMessage:
			err = c.writePreparedMessage(message)
		default:
			err = ErrInvalidMessage
		}
		if err != nil {
			c.handleWriteErr(c, err)
		}
	}
}

func (c *Conn[T]) writeMessage(message *Message) error {
	data := []byte(nil)
	if message.buffer != nil {
		data = message.buffer.Bytes()
	}
	if isControl(message.messageType) {
		return c.Conn.WriteControl(message.messageType, data, c.deadline(message.timeout))
	} else {
		c.Conn.SetWriteDeadline(c.deadline(message.timeout))
		return c.Conn.WriteMessage(message.messageType, data)
	}
}

func (c *Conn[T]) writePreparedMessage(message *websocket.PreparedMessage) error {
	c.Conn.SetWriteDeadline(c.deadline(c.defaultWriteTimeout))
	return c.Conn.WritePreparedMessage(message)
}
