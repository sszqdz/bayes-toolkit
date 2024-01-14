// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redislock

import (
	"crypto/sha1"
	"encoding/hex"
)

type RedisScripEnum int32

const (
	ScriptTryLock RedisScripEnum = iota
	ScriptUnlock
)

const (
	scriptTryLock string = `local tryLock = redis.call('SET', KEYS[1], '1', 'NX', 'EX', ARGV[1])
if (tryLock)
then
	redis.call('DEL', KEYS[2])
	return 1
end
tryLock = redis.call('RPOP', KEYS[2]) 
if (tryLock)
then
	return 1
end
return 0`

	scriptUnlock string = `local locking = redis.call('SET', KEYS[1], '1', 'XX', 'EX', ARGV[1])
if (locking)
then
    redis.call('LPUSH', KEYS[2], '1')
    redis.call('EXPIRE', KEYS[2], ARGV[1] + 10)
end`
)

var scriptDic = make(map[RedisScripEnum]string, 2)
var hashDic = make(map[RedisScripEnum]string, 2)

func init() {
	scriptDic[ScriptTryLock] = scriptTryLock
	scriptDic[ScriptUnlock] = scriptUnlock

	initHash()
}

func initHash() {
	for k, v := range scriptDic {
		h := sha1.New()
		h.Write([]byte(v))
		hashDic[k] = hex.EncodeToString(h.Sum(nil))
	}
}

func (enumCode RedisScripEnum) GetScript() string {
	return scriptDic[enumCode]
}

func (enumCode RedisScripEnum) GetHash() string {
	return hashDic[enumCode]
}
