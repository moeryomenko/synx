package synx

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkPool_BasicFunctionality(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkPool(ctx, 2)

	var counter int64
	for range 5 {
		err := pool.Go(func(_ context.Context) error {
			atomic.AddInt64(&counter, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if err := pool.Wait(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt64(&counter); got != 5 {
		t.Errorf("expected counter to be 5, got %d", got)
	}
}

func TestWorkPool_ConcurrencyLimit(t *testing.T) {
	ctx := context.Background()
	maxWorkers := int64(2)
	pool := NewWorkPool(ctx, maxWorkers)

	var concurrentCount, maxConcurrent int64

	for range 5 {
		err := pool.Go(func(_ context.Context) error {
			current := atomic.AddInt64(&concurrentCount, 1)
			for {
				max := atomic.LoadInt64(&maxConcurrent)
				if current <= max || atomic.CompareAndSwapInt64(&maxConcurrent, max, current) {
					break
				}
			}

			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&concurrentCount, -1)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if err := pool.Wait(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt64(&maxConcurrent); got > maxWorkers {
		t.Errorf("expected max concurrent workers to be <= %d, got %d", maxWorkers, got)
	}
}

func TestWorkPool_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkPool(ctx, 2)

	errTest := errors.New("test error")

	err := pool.Go(func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = pool.Go(func(_ context.Context) error {
		return errTest
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = pool.Wait()
	if err == nil {
		t.Error("expected error, got nil")
	}

	if !errors.Is(err, errTest) {
		t.Errorf("expected error to contain test error, got %v", err)
	}
}

func TestWorkPool_MultipleErrors(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkPool(ctx, 2)

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	err := pool.Go(func(_ context.Context) error {
		return err1
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = pool.Go(func(_ context.Context) error {
		return err2
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = pool.Wait()
	if err == nil {
		t.Error("expected error, got nil")
	}

	if !errors.Is(err, err1) || !errors.Is(err, err2) {
		t.Errorf("expected error to contain both errors, got %v", err)
	}
}

func TestWorkPool_TryGo(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkPool(ctx, 1)

	started := make(chan struct{})
	blocked := make(chan struct{})

	// Fill the pool
	success := pool.TryGo(func(_ context.Context) error {
		started <- struct{}{}
		<-blocked
		return nil
	})
	if !success {
		t.Error("expected TryGo to succeed when pool is empty")
	}

	// Wait for the first goroutine to start
	<-started

	// Try to add another - should fail
	success = pool.TryGo(func(_ context.Context) error {
		return nil
	})
	if success {
		t.Error("expected TryGo to fail when pool is full")
	}

	// Release the blocked goroutine
	close(blocked)

	if err := pool.Wait(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWorkPool_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewWorkPool(ctx, 2)

	// Cancel the context immediately
	cancel()

	err := pool.Go(func(_ context.Context) error {
		return nil
	})

	if err == nil {
		t.Error("expected error when context is cancelled")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestWorkPool_MaxWorkers(t *testing.T) {
	ctx := context.Background()
	maxWorkers := int64(5)
	pool := NewWorkPool(ctx, maxWorkers)

	if got := pool.MaxWorkers(); got != maxWorkers {
		t.Errorf("expected MaxWorkers() to return %d, got %d", maxWorkers, got)
	}
}

func TestWorkPool_Panic(t *testing.T) {
	ctx := context.Background()
	pool := NewWorkPool(ctx, 1)

	err := pool.Go(func(_ context.Context) error {
		panic("test panic")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = pool.Wait()
	if err == nil {
		t.Error("expected error from panic, got nil")
	}
}
