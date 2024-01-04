// References
// http.TimeoutHandler()
// https://github.com/vearne/gin-timeout
// https://github.com/gin-contrib/timeout

package gintimeout

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultTimeout = 3 * time.Second

func defaultResponse(gtx *gin.Context) {
	gtx.String(http.StatusRequestTimeout, http.StatusText(http.StatusRequestTimeout))
}

func onlyNextHandler(gtx *gin.Context) {
	gtx.Next()
}

// If you want to use ctx.Done() in your handler, remember that you have to set:
// engine.ContextWithFallback = true
func Timeout(opts ...Option) gin.HandlerFunc {
	to := &TimeoutOption{
		timeout:  defaultTimeout,
		response: defaultResponse,
	}
	for _, opt := range opts {
		if opt == nil {
			panic("timeout option not be nil")
		}
		opt(to)
	}
	if to.timeout <= 0 {
		return onlyNextHandler
	}
	bufPool := newBufPool(to.bufferSize)

	return func(gtx *gin.Context) {
		gtxCp := gtx.Dump()
		gtx.Abort()

		tw := &timeoutWriter{
			ResponseWriter: gtxCp.Writer,
			h:              make(http.Header),
			buf:            bufPool.getBuf(),
		}
		gtxCp.Writer = tw

		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(gtxCp.Request.Context(), to.timeout)
		defer cancel()
		gtxCp.Request = gtxCp.Request.WithContext(ctx)

		var finish atomic.Bool
		finish.Store(false)
		finishChan := make(chan struct{})
		panicChan := make(chan any, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					err := fmt.Errorf("gin-timeout recover:%v, stack: \n :%v", p, string(debug.Stack()))
					panicChan <- err
				}
			}()
			gtxCp.Next()
			finish.Store(true)
			close(finishChan)
		}()

		select {
		case p := <-panicChan:
			bufPool.putBuf(tw.buf)
			gtxCp.Abort()
			gtxCp.DumpTo(gtx)
			gtxCp.RecycleDump() // recycle the copy of Context
			panic(p)
		case <-ctx.Done():
			tw.mu.Lock()
			defer tw.mu.Unlock()
			// 再次探测正常业务逻辑是否已经走完，走完表示一定已经 write，因此走正常处理流程
			if finish.Load() {
				handleFinish(tw)
				bufPool.putBuf(tw.buf)
				gtxCp.Abort()
				gtxCp.RecycleDump() // recycle the copy of Context
				return
			}
			if tw.hijacked {
				return
			}

			gtxCp.Abort()
			gtxCp.DumpTo(gtx)
			to.response(gtx)
			tw.status = gtx.Writer.Status()
			tw.size = gtx.Writer.Size()
			tw.err = ctx.Err()
			// If timeout happen, the buffer and the copy of Context cannot be cleared actively,
			// but wait for the GC to recycle.
		case <-finishChan:
			tw.mu.Lock()
			defer tw.mu.Unlock()
			handleFinish(tw)
			bufPool.putBuf(tw.buf)
			gtxCp.DumpTo(gtx)
			gtxCp.Abort()
			gtxCp.RecycleDump() // recycle the copy of Context
		}
	}
}

func handleFinish(tw *timeoutWriter) {
	if tw.hijacked {
		return
	}
	// header
	dst := tw.ResponseWriter.Header()
	for k, vv := range tw.Header() {
		dst[k] = vv
	}
	// status
	if !tw.wroteHeader {
		tw.status = http.StatusOK
	}
	tw.ResponseWriter.WriteHeader(tw.status)
	// write
	if _, err := tw.ResponseWriter.Write(tw.buf.Bytes()); err != nil {
		panic(err)
	}
}
