// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redislock

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
)

// TODO safe lock

func HardLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	result, err := getClient().SetNX(ctx, key, 1, expiration).Result()
	if err != nil {
		return false, err
	}

	return result, nil
}

func ReleaseHardLock(ctx context.Context, key string) error {
	_, err := getClient().Del(ctx, key).Result()
	if err != nil {
		return err
	}

	return nil
}

func Lock(ctx context.Context, timeout time.Duration, key string, ex uint) error {
	listKey := key + listKeySuffix
	script := ScriptTryLock.GetScript()
	sha1 := ScriptTryLock.GetHash()
	result, err := getClient().EvalSha(ctx, sha1, []string{key, listKey}, ex).Result()
	if err != nil {
		if strings.HasPrefix(err.Error(), "NOSCRIPT") {
			result, err = getClient().Eval(ctx, script, []string{key, listKey}, ex).Result()
		}
		if err != nil {
			return err
		}
	}
	if cast.ToInt(result) == 1 {
		// Successfully acquired the lock.
		return nil
	}
	results, err := getClient().BRPop(ctx, timeout, listKey).Result()
	if err != nil {
		if err == redis.Nil { // timeout
			return errors.New("ACQUIRE LOCK ERR")
		} else {
			return err
		}
	}
	// Successfully acquired the lock.
	if len(results) == 2 && results[1] == "1" {
		return nil
	}

	return errors.New("ACQUIRE LOCK ERR")
}

func ReleaseLock(ctx context.Context, key string, ex uint) error {
	listKey := key + listKeySuffix
	script := ScriptUnlock.GetScript()
	sha1 := ScriptUnlock.GetHash()

	_, err := getClient().EvalSha(ctx, sha1, []string{key, listKey}, ex).Result()
	if err != nil {
		if strings.HasPrefix(err.Error(), "NOSCRIPT") {
			_, err = getClient().Eval(ctx, script, []string{key, listKey}, ex).Result()
		}
		if err == redis.Nil { // Successfully released the lock.
			return nil
		}
	}

	return errors.New("REDIS ERROR")
}
