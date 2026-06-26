package jobs

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolLimitsConcurrency(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	release := make(chan struct{})
	done := make(chan struct{}, 4)

	pool := NewPool(context.Background(), 2, 4, func(_ context.Context, _ string) {
		current := active.Add(1)
		for {
			previous := maximum.Load()
			if current <= previous || maximum.CompareAndSwap(previous, current) {
				break
			}
		}
		<-release
		active.Add(-1)
		done <- struct{}{}
	})
	t.Cleanup(pool.Stop)

	for _, id := range []string{"a", "b", "c", "d"} {
		if err := pool.Enqueue(id); err != nil {
			t.Fatalf("Enqueue(%q) error = %v", id, err)
		}
	}
	time.Sleep(50 * time.Millisecond)
	if got := maximum.Load(); got != 2 {
		t.Fatalf("maximum concurrency = %d, want 2", got)
	}
	close(release)
	for range 4 {
		<-done
	}
}

func TestPoolJobContextOutlivesRequestCancellation(t *testing.T) {
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	cancelRequest()

	var wg sync.WaitGroup
	wg.Add(1)
	processed := make(chan error, 1)
	pool := NewPool(context.Background(), 1, 1, func(ctx context.Context, _ string) {
		defer wg.Done()
		processed <- ctx.Err()
	})
	t.Cleanup(pool.Stop)

	if err := pool.EnqueueFrom(requestCtx, "asset"); err != nil {
		t.Fatalf("EnqueueFrom() error = %v", err)
	}
	wg.Wait()
	if err := <-processed; err != nil {
		t.Fatalf("worker context error = %v, want nil", err)
	}
}
