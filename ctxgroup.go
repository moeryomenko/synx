package synx

import (
	"context"
	"errors"
	"sync"
)

// A CtxGroup is a collection of goroutines working on subtasks that are part of
// the same overall task.
type CtxGroup struct {
	ctx context.Context

	wg *sync.WaitGroup

	lock sync.Mutex
	err  error
}

// NewCtxGroup returns a new CtxGroup with an derived Context from ctx.
func NewCtxGroup(ctx context.Context) *CtxGroup {
	return &CtxGroup{
		wg:  &sync.WaitGroup{},
		ctx: ctx,
	}
}

// Wait blocks until all function calls from the Go method have returned.
func (g *CtxGroup) Wait() error {
	g.wg.Wait()

	return g.err
}

// Go calls the given function in a new goroutine.
func (g *CtxGroup) Go(f func(ctx context.Context) error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := Graceful(g.ctx, f); err != nil {
			g.lock.Lock()
			g.err = errors.Join(g.err, err)
			g.lock.Unlock()
		}
	}()
}
