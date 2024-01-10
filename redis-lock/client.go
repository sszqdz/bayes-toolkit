package redislock

import "github.com/redis/go-redis/v9"

var client *redis.Client

func RegisterClient(lockClient *redis.Client) {
	client = lockClient
}

func getClient() *redis.Client {
	if client == nil {
		panic("nil client")
	}
	return client
}
