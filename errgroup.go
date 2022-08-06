package synx

import "context"

// A ErrGroup is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero ErrGroup is valid and does not cancel on error.
type ErrGroup struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg *WaitGroup

	errOnce *Once
	err     error
}

// NewErrGroup returns a new Group with an derived Context from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func NewErrGroup(ctx context.Context) *ErrGroup {
	ctx, cancel := context.WithCancel(ctx)
	return &ErrGroup{
		wg:      NewWaitGroup(),
		ctx:     ctx,
		cancel:  cancel,
		errOnce: new(Once),
	}
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *ErrGroup) Wait() error {
	<-g.wg.Wait()
	g.cancel()
	return g.err
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *ErrGroup) Go(f func(ctx context.Context) error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := Graceful(g.ctx, f); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	}()
}
