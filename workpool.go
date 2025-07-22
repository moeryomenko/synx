package synx

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/semaphore"
)

// A WorkPool manages a pool of goroutines with a maximum concurrency limit.
// It uses semaphore to control the number of concurrent operations.
type WorkPool struct {
	ctx        context.Context
	sem        *semaphore.Weighted
	maxWorkers int64

	lock sync.Mutex
	err  error
}

// NewWorkPool creates a new WorkPool with the specified maximum concurrency.
// maxWorkers defines the maximum number of concurrent goroutines that can run.
func NewWorkPool(ctx context.Context, maxWorkers int64) *WorkPool {
	return &WorkPool{
		ctx:        ctx,
		sem:        semaphore.NewWeighted(maxWorkers),
		maxWorkers: maxWorkers,
	}
}

// Go submits a function to be executed in the work pool.
// It will block if the maximum number of workers is already running.
// The function will be executed when a worker becomes available.
func (p *WorkPool) Go(f func(ctx context.Context) error) error {
	if err := p.sem.Acquire(p.ctx, 1); err != nil {
		return err
	}

	go func() {
		defer p.sem.Release(1)

		if err := Graceful(p.ctx, f); err != nil {
			p.lock.Lock()
			p.err = errors.Join(p.err, err)
			p.lock.Unlock()
		}
	}()

	return nil
}

// TryGo attempts to submit a function to be executed in the work pool.
// It returns false if the maximum number of workers is already running,
// true if the function was successfully submitted.
func (p *WorkPool) TryGo(f func(ctx context.Context) error) bool {
	if !p.sem.TryAcquire(1) {
		return false
	}

	go func() {
		defer p.sem.Release(1)

		if err := Graceful(p.ctx, f); err != nil {
			p.lock.Lock()
			p.err = errors.Join(p.err, err)
			p.lock.Unlock()
		}
	}()

	return true
}

// Wait blocks until all submitted work has completed and returns any errors
// that occurred during execution.
func (p *WorkPool) Wait() error {
	_ = p.sem.Acquire(p.ctx, p.maxWorkers)

	return p.err
}

// MaxWorkers returns the maximum number of concurrent workers.
func (p *WorkPool) MaxWorkers() int64 {
	return p.maxWorkers
}
