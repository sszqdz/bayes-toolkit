package redisqueue

import (
	"context"

	"github.com/adrianbrad/queue"
	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/conc"
)

type HandleFunc func(stream, key, val, msgId string) error
type IQueue interface {
	Info() *QueueInfo
	Handle(stream, key, val, msgId string) error
}

type QueueRunner struct {
	client    *redis.Client
	wgs       *queue.Linked[*conc.WaitGroup]
	closed    bool
	closeChan chan any
}

func NewQueueRunner(redisClient *redis.Client) *QueueRunner {
	if redisClient == nil {
		panic("nil client")
	}
	runner := &QueueRunner{
		client:    redisClient,
		wgs:       queue.NewLinked(make([]*conc.WaitGroup, 0)),
		closed:    false,
		closeChan: make(chan any),
	}

	return runner
}

func (qr *QueueRunner) Run(userQueues ...IQueue) error {
	for _, userQueue := range userQueues {
		if err := qr.run(userQueue); err != nil {
			return err
		}
	}
	return nil
}

func (qr *QueueRunner) run(userQueue IQueue) error {
	info := userQueue.Info()
	info.handler = userQueue.Handle
	checkQueueInfo(info)
	if err := qr.init(info); err != nil {
		return err
	}
	for i := 0; i < info.UserQueueInfo.ConsumerSize; i++ {
		info.wg.Go(func() { qr.normalRun(info) })
	}
	if !info.RetryQueueInfo.Stop {
		info.wg.Go(func() { qr.retryRun(info) })
	}
	if err := qr.wgs.Offer(info.wg); err != nil {
		return err
	}

	return nil
}

func (qr *QueueRunner) init(info *QueueInfo) error {
	ctx := context.Background()
	for _, stream := range info.UserQueueInfo.streams {
		_, err := qr.client.XGroupCreateMkStream(ctx, stream, info.UserQueueInfo.Group, info.UserQueueInfo.NewGroupStart).Result()
		if err != nil && !isBusyGroupErr(err) {
			return err
		}
	}
	return nil
}

func (qr *QueueRunner) Close() error {
	qr.closed = true
	close(qr.closeChan)
	_ = qr.client.Close()

	var firstErr error
	for wg := range qr.wgs.Iterator() {
		if err := wg.WaitAndRecover().AsError(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
