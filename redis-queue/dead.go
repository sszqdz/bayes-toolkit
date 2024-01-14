// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redisqueue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func (qr *QueueRunner) PageDead(ctx context.Context, stream, group string, cursor uint64, count int64) ([]string, uint64, error) {
	return qr.client.HScan(ctx, deadHashMapName(stream, group), cursor, "", count).Result()
}

func (qr *QueueRunner) HandleDead(ctx context.Context, handler HandleFunc, stream, group string, ids ...string) []error {
	if len(ids) == 0 {
		return nil
	}
	errs := make([]error, 0)
	deadName := deadHashMapName(stream, group)
	msgIfcs, err := qr.client.HMGet(ctx, deadName, ids...).Result()
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	succedlIds := make([]string, 0, len(ids))
	for i, msgIfc := range msgIfcs {
		if msgIfc == nil {
			succedlIds = append(succedlIds, ids[i])
			continue
		}
		var xMessage *redis.XMessage
		if err = json.UnmarshalFromString(msgIfc.(string), &xMessage); err != nil {
			errs = append(errs, err)
			continue
		}
		if xMessage == nil {
			errs = append(errs, err)
			continue
		}
		key, val := extractValues(xMessage.Values)
		if err := handler(stream, key, val, xMessage.ID); err != nil {
			errs = append(errs, err)
			continue
		}
		succedlIds = append(succedlIds, xMessage.ID)
	}
	if len(succedlIds) == 0 {
		return errs
	}

	_, err = qr.client.HDel(ctx, deadName, succedlIds...).Result()
	errs = append(errs, err)
	return errs
}

func (qr *QueueRunner) CleanDead(ctx context.Context, stream, group string, start, batchSize int64) error {
	_, err := qr.client.Del(ctx, deadHashMapName(stream, group)).Result()
	return err
}
