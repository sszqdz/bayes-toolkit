// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redisqueue

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/redis/go-redis/v9"
)

func (qr *QueueRunner) normalRun(info *QueueInfo) {
	defer func() {
		if r := recover(); r != nil {
			info.NotifyPanic(r, string(debug.Stack()))
			panic(r)
		}
	}()

	ctx := context.Background()
	for {
		xStreams, err := qr.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    info.UserQueueInfo.Group,
			Consumer: info.UserQueueInfo.consumer,
			Streams:  info.UserQueueInfo.Streams,
			Count:    info.UserQueueInfo.BatchSize,
			Block:    0,
			NoAck:    false,
		}).Result()
		if qr.closed {
			break
		}
		if err != nil {
			if err == redis.ErrClosed {
				continue
			}
			if err != redis.Nil {
				info.NotifyErr("", "", err)
			}
			time.Sleep(time.Second)
			continue
		}
		for _, xStream := range xStreams {
			stream := xStream.Stream
			for _, xMessage := range xStream.Messages {
				qr.handleMessage(ctx, info, stream, xMessage)
			}
		}
	}
}

func (qr *QueueRunner) handleMessage(ctx context.Context, info *QueueInfo, stream string, xMessage redis.XMessage) {
	key, val := extractValues(xMessage.Values)
	if err := info.handler(stream, key, val, xMessage.ID); err != nil {
		info.NotifyErr(stream, key, err)
		return
	}

	if _, err := qr.client.XAck(ctx, stream, info.UserQueueInfo.Group, xMessage.ID).Result(); err != nil {
		info.NotifyErr(stream, key, err)
		return
	}
}
