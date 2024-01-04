// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gintimeout

import (
	"bytes"
	"sync"
)

const smallBufferSize = 64

type bufPool struct {
	pool    sync.Pool
	size    int
	maxSize int
}

func newBufPool(size int) *bufPool {
	if size == 0 {
		size = 4 * 1024
	} else if size < smallBufferSize {
		size = smallBufferSize
	}

	return &bufPool{
		pool: sync.Pool{New: func() any {
			return bytes.NewBuffer(make([]byte, 0, size))
		}},
		size:    size,
		maxSize: size << 2,
	}
}

func (p *bufPool) getBuf() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *bufPool) putBuf(buffer *bytes.Buffer) {
	cap := buffer.Cap()
	if cap <= p.maxSize && cap >= p.size { // Prevents oversize and undersize []byte pooling
		buffer.Reset()
		p.pool.Put(buffer)
	}
}
