// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redislock

import "github.com/redis/go-redis/v9"

var (
	client        *redis.Client
	listKeySuffix = "-locklist"
)

func RegisterClient(redisClient *redis.Client, suffix ...string) {
	if redisClient == nil {
		panic("nil client")
	}
	client = redisClient
	if len(suffix) > 0 {
		listKeySuffix = suffix[0]
	}
}

func getClient() *redis.Client {
	if client == nil {
		panic("nil client")
	}
	return client
}
