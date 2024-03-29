package rrand

import (
	"math/rand/v2"
	"time"
)

func RandDuration(maxDura int64, duraUnit time.Duration) time.Duration {
	return time.Duration(rand.Int64N(maxDura)) * duraUnit
}
