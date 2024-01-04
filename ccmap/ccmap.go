package ccmap

import (
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/spf13/cast"
)

func New[K cInteger, V any]() cmap.ConcurrentMap[K, V] {
	return cmap.NewWithCustomShardingFunction[K, V](strfnv32[K])
}

type cInteger interface {
	uint | uint8 | uint16 | uint32 | uint64 |
		int | int8 | int16 | int32 | int64
}

func strfnv32[K cInteger](key K) uint32 {
	return fnv32(cast.ToString(key))
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
