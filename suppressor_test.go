package synx

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoDeduplicate(t *testing.T) {
	g := NewSuppressor(time.Second)
	var calls int32
	fn := func() (interface{}, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(15 * time.Millisecond)

		return `test`, nil
	}

	ticker := time.NewTicker(10 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

loop:
	for {
		select {
		case <-ticker.C:
			_ = <-g.Do(`test`, fn)
		case <-ctx.Done():
			break loop
		}
	}

	if calls != 1 {
		t.Errorf("unexpected calls count: %d", calls)
	}
}
