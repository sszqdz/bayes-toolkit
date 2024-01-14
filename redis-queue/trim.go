// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redisqueue

import (
	"context"
	"time"

	"github.com/sourcegraph/conc"
	"github.com/spf13/cast"
)

type TrimInfo struct {
	Streams     []string
	Tick        time.Duration
	MaxLen      int64
	MaxDuration time.Duration
	Limit       int64
}

func (qr *QueueRunner) RunTrim(trimInfos ...*TrimInfo) error {
	wg := conc.NewWaitGroup()
	for _, trimInfo := range trimInfos {
		checkTrimInfo(trimInfo)
		wg.Go(func() {
			ctx := context.Background()
			tick := time.NewTicker(trimInfo.Tick)
			for {
				select {
				case <-qr.closeChan:
					tick.Stop()
					return
				case <-tick.C:
					qr.trimRun(ctx, trimInfo)
				}
			}
		})
	}
	if err := qr.wgs.Offer(wg); err != nil {
		return err
	}

	return nil
}

func checkTrimInfo(info *TrimInfo) {
	if info == nil {
		panic("nil trim info")
	}
	if len(info.Streams) == 0 {
		panic("empty streams")
	}
	for _, stream := range info.Streams {
		if stream == "" {
			panic("empty stream")
		}
	}
	if info.Tick < 0 {
		panic("invalid tick")
	}
	if info.Tick == 0 {
		info.Tick = time.Hour
	}
	if info.MaxLen <= 0 && info.MaxDuration <= 0 {
		panic("must have one strategy")
	}
	if info.Limit < 0 {
		panic("invalid limit")
	}
	if info.Limit == 0 {
		info.Limit = 100
	}
}

func (qr *QueueRunner) trimRun(ctx context.Context, trimInfo *TrimInfo) {
	for _, stream := range trimInfo.Streams {
		if trimInfo.MaxLen > 0 {
			qr.client.XTrimMaxLenApprox(ctx, stream, trimInfo.MaxLen, trimInfo.Limit)
		}
		if trimInfo.MaxDuration > 0 {
			minId := cast.ToString(time.Now().Add(-trimInfo.MaxDuration).UnixMilli())
			qr.client.XTrimMinIDApprox(ctx, stream, minId, trimInfo.Limit)
		}
	}
}
