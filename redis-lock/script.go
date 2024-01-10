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
	scriptTryLock string = `local tryLock = redis.call('set', KEYS[1], '1', 'ex', ARGV[1], 'nx')
if (tryLock)
then
	redis.call('del', KEYS[2])
	return 1
end
tryLock = redis.call('rpop', KEYS[2]) 
if (tryLock)
then
	return 1
end
return 0`

	scriptUnlock string = `local locking = redis.call('set', KEYS[1], '1', 'ex', ARGV[1], 'xx')
if (locking)
then
    redis.call('lpush', KEYS[2], '1')
    redis.call('expire', KEYS[2], ARGV[1] + 10)
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
