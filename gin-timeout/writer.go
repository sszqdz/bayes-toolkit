package gintimeout

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type timeoutWriter struct {
	gin.ResponseWriter
	h   http.Header
	buf *bytes.Buffer

	mu          sync.Mutex
	hijacked    bool
	wroteHeader bool
	status      int
	size        int
	err         error
}

func (tw *timeoutWriter) Header() http.Header {
	return tw.h
}

func (tw *timeoutWriter) Write(data []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.err != nil {
		return 0, tw.err
	}
	if !tw.wroteHeader {
		tw.writeHeader(http.StatusOK)
	}
	tw.size += len(data)

	return tw.buf.Write(data)
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.writeHeader(code)
}

func (tw *timeoutWriter) writeHeader(code int) {
	if tw.err != nil {
		return
	}
	if code == -1 { // gin is using -1 to skip writing the status code
		return
	}
	tw.wroteHeader = true
	tw.status = code
}

func (tw *timeoutWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	conn, rw, err := tw.ResponseWriter.(http.Hijacker).Hijack()
	if err == nil {
		tw.hijacked = true
		tw.err = http.ErrHijacked
	}

	return conn, rw, err
}

func (tw *timeoutWriter) Status() int {
	return tw.status
}

func (tw *timeoutWriter) Size() int {
	return tw.size
}

func (w *timeoutWriter) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

// !!! don't override
// func (tw *timeoutWriter) Written() bool

func (tw *timeoutWriter) WriteHeaderNow() {
	// ignore
}
