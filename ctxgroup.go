package synx

import (
	"context"
	"errors"
)

// A CtxGroup is a collection of goroutines working on subtasks that are part of
// the same overall task.
type CtxGroup struct {
	ctx context.Context

	wg *WaitGroup

	lock Spinlock
	err  error
}

// NewCtxGroup returns a new CtxGroup with an derived Context from ctx.
func NewCtxGroup(ctx context.Context) *CtxGroup {
	return &CtxGroup{
		wg:  NewWaitGroup(),
		ctx: ctx,
	}
}

// Wait blocks until all function calls from the Go method have returned.
func (g *CtxGroup) Wait() error {
	select {
	case <-g.ctx.Done():
		return g.ctx.Err()
	case <-g.wg.Wait():
		return g.err
	}
}

// Go calls the given function in a new goroutine.
func (g *CtxGroup) Go(f func(ctx context.Context) error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := Graceful(g.ctx, f); err != nil {
			g.lock.Lock()
                        err = errors.Join(err)
			g.lock.Unlock()
		}
	}()
}
