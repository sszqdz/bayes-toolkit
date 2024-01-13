package redisqueue

import (
	"time"

	"github.com/sourcegraph/conc"
)

type QueueInfo struct {
	UserQueueInfo  *UserQueueInfo
	RetryQueueInfo *RetryQueueInfo
	DeadQueueInfo  *DeadQueueInfo
	NotifyErr      func(stream, key string, err error)
	NotifyPanic    func(pnc any, stack string)

	wg      *conc.WaitGroup
	handler HandleFunc
}

type UserQueueInfo struct {
	Streams       []string
	Group         string
	ConsumerSize  int
	BatchSize     int64
	NewGroupStart string

	streams  []string // only stream names
	consumer string
}

type RetryQueueInfo struct {
	Stop        bool
	Tick        time.Duration
	MinRetry    int64
	MinIdleTime time.Duration
	BatchSize   int64
	NotifyDead  func(stream, key, val, msgId string)

	consumer string
}

type DeadQueueInfo struct {
	Stop bool
}

func checkQueueInfo(info *QueueInfo) {
	if info == nil || info.UserQueueInfo == nil {
		panic("nil queue info")
	}
	userQueueInfo := info.UserQueueInfo
	if len(userQueueInfo.Streams) == 0 {
		panic("empty streams")
	}
	if len(userQueueInfo.Streams)%2 != 0 {
		panic("invalid streams")
	}
	for _, stream := range userQueueInfo.Streams {
		if stream == "" {
			panic("empty stream")
		}
	}
	if userQueueInfo.Group == "" {
		panic("empty group")
	}
	if userQueueInfo.ConsumerSize < 0 {
		panic("invalid consumer size")
	}
	if userQueueInfo.ConsumerSize == 0 {
		userQueueInfo.ConsumerSize = 1
	}
	if userQueueInfo.BatchSize < 0 {
		panic("invalid batch size")
	}
	if userQueueInfo.BatchSize == 0 {
		userQueueInfo.BatchSize = 1
	}
	if userQueueInfo.NewGroupStart == "" {
		userQueueInfo.NewGroupStart = "$"
	}
	userQueueInfo.streams = extractStreamNames(userQueueInfo.Streams)
	userQueueInfo.consumer = userQueueInfo.Group + "-consumer"
	if info.RetryQueueInfo == nil {
		info.RetryQueueInfo = &RetryQueueInfo{}
	}
	retryQueueInfo := info.RetryQueueInfo
	if !retryQueueInfo.Stop {
		if retryQueueInfo.Tick < 0 {
			panic("invalid tick")
		}
		if retryQueueInfo.Tick == 0 {
			retryQueueInfo.Tick = time.Minute
		}
		if retryQueueInfo.MinRetry < 0 {
			panic("invalid max retry")
		}
		if retryQueueInfo.MinRetry == 0 {
			retryQueueInfo.MinRetry = 2
		}
		if retryQueueInfo.MinIdleTime < 0 {
			panic("invalid min idle time")
		}
		if retryQueueInfo.MinIdleTime == 0 {
			retryQueueInfo.MinIdleTime = 3 * time.Minute
		}
		if retryQueueInfo.BatchSize < 0 {
			panic("invalid batch size")
		}
		if retryQueueInfo.BatchSize == 0 {
			retryQueueInfo.BatchSize = 10
		}
		if retryQueueInfo.NotifyDead == nil {
			retryQueueInfo.NotifyDead = func(stream, key, val, msgId string) {}
		}
	}
	retryQueueInfo.consumer = userQueueInfo.consumer + "-retry"
	if info.DeadQueueInfo == nil {
		info.DeadQueueInfo = &DeadQueueInfo{}
	}
	if info.NotifyErr == nil {
		info.NotifyErr = func(stream, key string, err error) {}
	}
	if info.NotifyPanic == nil {
		info.NotifyPanic = func(pnc any, stack string) {}
	}
	info.wg = conc.NewWaitGroup()
}
