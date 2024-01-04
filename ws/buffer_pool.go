package ws

import (
	"bytes"
	"sync"
)

var (
	poolUsed          = false
	defaultBufferSize = bytes.MinRead
	maxBufferSize     = defaultBufferSize << 1
	bufferPool        *sync.Pool
)

func UseSmartBufferPool(bufferSize int) {
	poolUsed = true
	if bufferSize > bytes.MinRead {
		defaultBufferSize = bufferSize
		maxBufferSize = bufferSize << 1
	}
	bufferPool = &sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
		},
	}
}

func GetWriteBuffer() *bytes.Buffer {
	if !poolUsed {
		panic("You must first call UseBufferPool function once")
	}
	return getBuffer()
}

func PutReadBuffer(buffer *bytes.Buffer) {
	if !poolUsed {
		panic("You must first call UseBufferPool function once")
	}
	putBuffer(buffer)
}

func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func putBuffer(buffer *bytes.Buffer) {
	if poolUsed && buffer != nil {
		cap := buffer.Cap()
		if cap <= maxBufferSize && cap >= defaultBufferSize { // 防止超大和过小 []byte 入池
			buffer.Reset()
			bufferPool.Put(buffer)
		}
	}
}
