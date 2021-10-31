package synx

import (
	"context"
	"time"
)

// Schedule calls function `f` with a period `p` offsetted by `o`.
func Schedule(ctx context.Context, period time.Duration, offset time.Duration, f func()) {
	// Position the first execution.
	first := time.Now().Truncate(period).Add(offset)
	if first.Before(time.Now()) {
		first = first.Add(period)
	}
	firstC := time.After(first.Sub(time.Now()))

	// Receiving from a nil channel blocks forever.
	t := &time.Ticker{C: nil}

	for {
		select {
		case <-firstC:
			// The ticker has to be started before
			// function as it can take some time to finish.
			t = time.NewTicker(period)
			f()
		case <-t.C:
			f()
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}
