package synx

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidTimeout         = errors.New("timeout is zero or negative")
	ErrContextWithoutDeadline = errors.New("context has no deadline")
)

func CallWithContext(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := ctx.Deadline(); !ok {
		return ErrContextWithoutDeadline
	}
	result := make(chan error, 1)

	go func() {
		defer close(result)
		err := Graceful(ctx, fn)
		result <- err
	}()

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func CallWithTimeout(timeout time.Duration, fn func(context.Context) error) error {
	if timeout <= 0 {
		return ErrInvalidTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return CallWithContext(ctx, fn)
}
