package redisqueue

import (
	"context"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

type retryQueueData struct {
	deadIds                []string
	retryIds               []string
	sendToDeadIds          []string
	sendToDeadXMessagesMap map[string]any

	running atomic.Bool
}

func (data *retryQueueData) reset() {
	clear(data.deadIds)
	data.deadIds = data.deadIds[:0]
	clear(data.retryIds)
	data.retryIds = data.retryIds[:0]
	clear(data.sendToDeadIds)
	data.sendToDeadIds = data.sendToDeadIds[:0]
	clear(data.sendToDeadXMessagesMap)
}

func (qr *QueueRunner) retryRun(info *QueueInfo) {
	defer func() {
		if r := recover(); r != nil {
			info.NotifyPanic(r, string(debug.Stack()))
			panic(r)
		}
	}()

	ctx := context.Background()
	data := &retryQueueData{
		deadIds:                make([]string, 0, info.RetryQueueInfo.BatchSize),
		retryIds:               make([]string, 0, info.RetryQueueInfo.BatchSize),
		sendToDeadIds:          make([]string, 0, info.RetryQueueInfo.BatchSize),
		sendToDeadXMessagesMap: make(map[string]any, 0),
	}
	data.running.Store(false)

	tick := time.NewTicker(info.RetryQueueInfo.Tick)
	for {
		select {
		case <-qr.closeChan:
			tick.Stop()
			return
		case <-tick.C:
			qr.retryHandle(ctx, info, data)
		}
	}
}

func (qr *QueueRunner) retryHandle(ctx context.Context, info *QueueInfo, data *retryQueueData) {
	if !data.running.CompareAndSwap(false, true) {
		return
	}
	defer data.running.Store(false)

	for _, stream := range info.UserQueueInfo.streams {
		data.reset()
		xPendingExts, err := qr.client.XPendingExt(ctx, &redis.XPendingExtArgs{
			Stream: stream,
			Group:  info.UserQueueInfo.Group,
			Idle:   info.RetryQueueInfo.MinIdleTime,
			Start:  "-",
			End:    "+",
			Count:  info.RetryQueueInfo.BatchSize,
		}).Result()
		if qr.closed {
			break
		}
		if err != nil {
			if err == redis.ErrClosed {
				continue
			}
			info.NotifyErr("", "", err)
			continue
		}
		for _, xPendingExt := range xPendingExts {
			if xPendingExt.RetryCount > info.RetryQueueInfo.MinRetry {
				data.deadIds = append(data.deadIds, xPendingExt.ID)
				continue
			}
			data.retryIds = append(data.retryIds, xPendingExt.ID)
		}
		// Messages that have been deleted do not appear in the XClaim return result
		qr.claimAndSendToDead(ctx, stream, data, info)
		qr.claimAndRetry(ctx, stream, data, info)
	}
}

func (qr *QueueRunner) claimAndSendToDead(ctx context.Context, stream string, data *retryQueueData, info *QueueInfo) {
	if len(data.deadIds) == 0 {
		return
	}
	xMessages, err := qr.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   stream,
		Group:    info.UserQueueInfo.Group,
		Consumer: info.RetryQueueInfo.consumer,
		MinIdle:  info.RetryQueueInfo.MinIdleTime,
		Messages: data.deadIds,
	}).Result()
	if qr.closed {
		return
	}
	if err != nil {
		if err == redis.ErrClosed {
			return
		}
		info.NotifyErr("", "", err)
		return
	}
	if len(xMessages) == 0 {
		return
	}

	for _, xMessage := range xMessages {
		data.sendToDeadIds = append(data.sendToDeadIds, xMessage.ID)
		xMessageStr, _ := json.MarshalToString(xMessage)
		data.sendToDeadXMessagesMap[xMessage.ID] = xMessageStr
	}
	if !info.DeadQueueInfo.Stop {
		// send to dead hash
		if _, err := qr.client.HSet(ctx, deadHashMapName(stream, info.UserQueueInfo.Group), data.sendToDeadXMessagesMap).Result(); err != nil {
			if err == redis.ErrClosed {
				return
			}
			info.NotifyErr(stream, "", err)
			return
		}
	}
	if _, err = qr.client.XAck(ctx, stream, info.UserQueueInfo.Group, data.sendToDeadIds...).Result(); err != nil {
		info.NotifyErr(stream, "", err)
		return
	}
	for _, xMessage := range xMessages {
		key, val := extractValues(xMessage.Values)
		info.RetryQueueInfo.NotifyDead(stream, key, val, xMessage.ID)
	}
}

func (qr *QueueRunner) claimAndRetry(ctx context.Context, stream string, data *retryQueueData, info *QueueInfo) {
	if len(data.retryIds) == 0 {
		return
	}
	xMessages, err := qr.client.XClaim(ctx, &redis.XClaimArgs{
		Stream:   stream,
		Group:    info.UserQueueInfo.Group,
		Consumer: info.RetryQueueInfo.consumer,
		MinIdle:  info.RetryQueueInfo.MinIdleTime,
		Messages: data.retryIds,
	}).Result()
	if qr.closed {
		return
	}
	if err != nil {
		if err == redis.ErrClosed {
			return
		}
		info.NotifyErr("", "", err)
		return
	}
	for _, xMessage := range xMessages {
		key, val := extractValues(xMessage.Values)
		if err := info.handler(stream, key, val, xMessage.ID); err != nil {
			info.NotifyErr(stream, key, err)
			continue
		}
		if _, err = qr.client.XAck(ctx, stream, info.UserQueueInfo.Group, xMessage.ID).Result(); err != nil {
			info.NotifyErr(stream, key, err)
			continue
		}
	}
}
