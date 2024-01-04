package ws

import (
	"bytes"
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrInvalidTimeout = errors.New("invalid timeout param")
	ErrInvalidMessage = errors.New("invalid mesaage")
	ErrNilMessage     = errors.New("nil message")

	closeNormalMessage = websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
)

type Conn[T any] struct {
	writeChan    chan any
	writeEndChan chan struct{}
	closeChan    chan struct{}
	isClosing    atomic.Bool
	isClosed     atomic.Bool
	mutex        sync.Mutex

	closeTimeout        time.Duration
	defaultWriteTimeout time.Duration
	handleMessage       func(conn *Conn[T], messageType int, buffer *bytes.Buffer)
	handleMessageErr    func(conn *Conn[T], normal bool, err error)
	handleWriteErr      func(conn *Conn[T], err error)
	onClosingHandle     func(conn *Conn[T])

	Conn       *websocket.Conn
	Identifier T
}

func WrapConn[T any](identifier T, conn *websocket.Conn, writeChanSize int) *Conn[T] {
	c := &Conn[T]{
		writeChan:    make(chan any, writeChanSize),
		writeEndChan: make(chan struct{}),
		closeChan:    make(chan struct{}),

		Conn:       conn,
		Identifier: identifier,
	}
	c.isClosing.Store(false)
	c.isClosed.Store(false)

	c.SetCloseTimeout(time.Second)
	c.SetDefaultWriteTimeout(time.Second)
	c.SetMessageHandler(nil)
	c.SetMessageErrHandler(nil)
	c.SetWriteErrHandler(nil)
	c.SetOnClosingHandler(nil)

	// Take over SetCloseMessageHandler, SetPingMessageHandler of gorilla/websocket
	c.SetCloseMessageHandler(time.Second, nil)
	c.SetPingMessageHandler(time.Second, nil)

	go c.readLoop()
	go c.writeLoop()

	return c
}

func (c *Conn[T]) SetCloseTimeout(timeout time.Duration) {
	c.closeTimeout = timeout
}

func (c *Conn[T]) SetDefaultWriteTimeout(timeout time.Duration) {
	c.defaultWriteTimeout = timeout
}

func (c *Conn[T]) SetMessageHandler(h func(conn *Conn[T], messageType int, buffer *bytes.Buffer)) {
	if h == nil {
		h = func(conn *Conn[T], messageType int, buffer *bytes.Buffer) {}
	}
	c.handleMessage = h
}

func (c *Conn[T]) SetMessageErrHandler(h func(conn *Conn[T], normal bool, err error)) {
	if h == nil {
		h = func(conn *Conn[T], normal bool, err error) {}
	}
	c.handleMessageErr = h
}

func (c *Conn[T]) SetWriteErrHandler(h func(conn *Conn[T], err error)) {
	if h == nil {
		h = func(conn *Conn[T], err error) {}
	}
	c.handleWriteErr = h
}

func (c *Conn[T]) SetOnClosingHandler(h func(conn *Conn[T])) {
	if h == nil {
		h = func(conn *Conn[T]) {}
	}
	c.onClosingHandle = h
}

func (c *Conn[T]) SetPingMessageHandler(timeout time.Duration, h func(message string) error) {
	if h == nil {
		h = func(message string) error {
			err := c.Conn.WriteControl(websocket.PongMessage, []byte(message), c.deadline(timeout))
			if err == websocket.ErrCloseSent {
				return nil
			} else if _, ok := err.(net.Error); ok {
				return nil
			}
			return err
		}
	}
	c.Conn.SetPingHandler(h)
}

func (c *Conn[T]) SetCloseMessageHandler(timeout time.Duration, h func(code int, text string) error) {
	if h == nil {
		h = func(code int, text string) error {
			return c.Close(code, text)
		}
	}
	c.Conn.SetCloseHandler(h)
}

func (c *Conn[T]) Shutdown() {
	if c.isClosed.Load() {
		return
	}
	// set closing flag
	c.isClosing.Store(true)

	c.mutex.Lock()
	if !c.isClosed.Load() {
		// A channel can only be closed once, ensuring that this code is only executed once
		close(c.closeChan)
		close(c.writeChan)
		// send close message
		_ = c.writeMessage(&Message{messageType: websocket.CloseMessage, buffer: bytes.NewBuffer(closeNormalMessage), timeout: c.defaultWriteTimeout})
		// close conn
		_ = c.Conn.Close()
		// set closed flag
		c.isClosed.Store(true)
	}
	c.mutex.Unlock()
}

func (c *Conn[T]) Close(closeCode int, text ...string) error {
	var err error
	if c.isClosed.Load() {
		return err
	}
	// set closing flag
	c.isClosing.Store(true)

	code := websocket.CloseNormalClosure
	if closeCode != 0 {
		code = closeCode
	}
	txt := ""
	switch ms := len(text); ms {
	case 0:
	case 1:
		txt = text[0]
	default:
		var sb strings.Builder
		for i, textPiece := range text {
			if i != 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(textPiece)
		}
		txt = sb.String()
	}
	message := websocket.FormatCloseMessage(code, txt)

	c.mutex.Lock()
	if !c.isClosed.Load() {
		// A channel can only be closed once, ensuring that this code is only executed once
		close(c.closeChan)
		close(c.writeChan)
		// wait for writing remaining messages
		select {
		case <-time.After(c.closeTimeout):
		case <-c.writeEndChan:
		}
		// send close message
		_ = c.writeMessage(&Message{messageType: websocket.CloseMessage, buffer: bytes.NewBuffer(message), timeout: c.defaultWriteTimeout})
		// close conn
		err = c.Conn.Close()
		// set flag
		c.isClosed.Store(true)
		// handle closing callback
		c.onClosingHandle(c)
	}
	c.mutex.Unlock()

	return err
}

func (c *Conn[T]) Write(messageType int, buffer *bytes.Buffer, timeout ...time.Duration) error {
	ps := len(timeout)
	if ps > 1 {
		return ErrInvalidTimeout
	}
	to := c.defaultWriteTimeout
	if ps == 1 {
		to = timeout[0]
	}
	return c.sendWriteChannel(&Message{messageType: messageType, buffer: buffer, timeout: to})
}

func (c *Conn[T]) WriteMessage(message *Message) error {
	if message == nil {
		return ErrNilMessage
	}
	return c.sendWriteChannel(message)
}

func (c *Conn[T]) WritePreparedMessage(message *websocket.PreparedMessage) error {
	if message == nil {
		return ErrNilMessage
	}
	return c.sendWriteChannel(message)
}

func (c *Conn[T]) WriteControl(messageType int, buffer *bytes.Buffer, timeout time.Duration) error {
	return c.sendWriteChannel(&Message{messageType: messageType, buffer: buffer, timeout: timeout})
}

func (c *Conn[T]) sendWriteChannel(message any) (err error) {
	if c.isClosing.Load() {
		// !!! Prevent writing to closed channel without mutex
		// This approach does not completely prevent writing to closed channel, but it has higher performance
		return websocket.ErrCloseSent
	}
	defer func() {
		if p := recover(); p != nil {
			err = websocket.ErrCloseSent
		}
	}()
	select {
	case <-c.closeChan: // prevent blocking
		return websocket.ErrCloseSent
	case c.writeChan <- message:
		return nil
	}
}

func (c *Conn[T]) deadline(timeout time.Duration) time.Time {
	ddl := time.Time{}
	if timeout == 0 {
		timeout = c.defaultWriteTimeout
	}
	if timeout != 0 {
		ddl = time.Now().Add(timeout)
	}
	return ddl
}
