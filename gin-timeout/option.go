// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gintimeout

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Option func(*TimeoutOption)

type TimeoutOption struct {
	timeout    time.Duration
	response   gin.HandlerFunc
	bufferSize int
}

func WithTimeout(timeout time.Duration) Option {
	return func(t *TimeoutOption) {
		t.timeout = timeout
	}
}

func WithResponse(handler gin.HandlerFunc) Option {
	return func(t *TimeoutOption) {
		t.response = handler
	}
}

func WithBufferSize(size int) Option {
	return func(t *TimeoutOption) {
		t.bufferSize = size
	}
}
