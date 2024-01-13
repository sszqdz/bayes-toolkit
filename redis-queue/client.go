package redisqueue

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

func (qr *QueueRunner) Send(ctx context.Context, stream, val string) (string, error) {
	if stream == "" || val == "" {
		return "", errors.New("invalid params")
	}

	return qr.client.XAdd(ctx, &redis.XAddArgs{
		Stream:     stream,
		NoMkStream: true,
		ID:         "*",
		Values:     map[string]interface{}{valStr: val},
	}).Result()
}

func (qr *QueueRunner) SendWithKey(ctx context.Context, stream, key, val string) (string, error) {
	if stream == "" || key == "" || val == "" {
		return "", errors.New("invalid params")
	}

	return qr.client.XAdd(ctx, &redis.XAddArgs{
		Stream:     stream,
		NoMkStream: true,
		ID:         "*",
		Values: map[string]interface{}{
			keyStr: key,
			valStr: val,
		},
	}).Result()
}
