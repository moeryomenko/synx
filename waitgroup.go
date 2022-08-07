package synx

import "sync/atomic"

// WaitGroup is like sync.WaitGroup with a signal channel.
type WaitGroup struct {
	n      int64
	doneCh chan struct{}
}

// NewWaitGroup returns a new WaitGroup.
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		doneCh: make(chan struct{}),
	}
}

// Add has same behaviour as sync.WaitGroup.
func (wg *WaitGroup) Add(delta int) {
	if n := atomic.AddInt64(&wg.n, int64(delta)); n == 0 {
		if !isChanClosed(wg.doneCh) {
			close(wg.doneCh)
		}
	}
}

// Done has same behaviour as sync.WaitGroup.
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

// Wait returns a channel that will be closed on completion.
func (wg *WaitGroup) Wait() <-chan struct{} {
	return wg.doneCh
}
