package jobs

import (
	"context"
	"errors"
	"sync"
)

var ErrQueueFull = errors.New("asset job queue is full")

type Pool struct {
	cancel context.CancelFunc
	queue  chan string
	wg     sync.WaitGroup
	once   sync.Once
}

func NewPool(
	parent context.Context,
	workers int,
	queueSize int,
	handler func(context.Context, string),
) *Pool {
	if workers < 1 {
		workers = 1
	}
	if queueSize < 1 {
		queueSize = workers
	}
	ctx, cancel := context.WithCancel(parent)
	pool := &Pool{
		cancel: cancel,
		queue:  make(chan string, queueSize),
	}
	for range workers {
		pool.wg.Add(1)
		go func() {
			defer pool.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case assetID := <-pool.queue:
					handler(ctx, assetID)
				}
			}
		}()
	}
	return pool
}

func (p *Pool) Enqueue(assetID string) error {
	select {
	case p.queue <- assetID:
		return nil
	default:
		return ErrQueueFull
	}
}

func (p *Pool) EnqueueFrom(_ context.Context, assetID string) error {
	return p.Enqueue(assetID)
}

func (p *Pool) Stop() {
	p.once.Do(func() {
		p.cancel()
		p.wg.Wait()
	})
}
