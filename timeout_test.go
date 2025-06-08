package synx_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/moeryomenko/synx"
)

func Test_CallWithContext(t *testing.T) {
	err := synx.CallWithContext(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if !errors.Is(err, synx.ErrContextWithoutDeadline) {
		t.Errorf("unexpected error: %v, want: %v", err, synx.ErrContextWithoutDeadline)
	}
}

func Test_CallWithTimeout(t *testing.T) {
	t.Parallel()

	errFailed := errors.New("failed")

	testcases := map[string]struct {
		timeout time.Duration
		fn      func(context.Context) error
		err     error
	}{
		"function compeletes before timeout without error": {
			timeout: 2 * time.Second,
			fn: func(ctx context.Context) error {
				return nil
			},
			err: nil,
		},
		"function compeletes before timeout with error": {
			timeout: 2 * time.Second,
			fn: func(ctx context.Context) error {
				return errFailed
			},
			err: errFailed,
		},
		"timeout is zero": {
			timeout: 0,
			fn: func(ctx context.Context) error {
				return nil
			},
			err: synx.ErrInvalidTimeout,
		},
		"context timeout triggered": {
			timeout: 10 * time.Millisecond,
			fn: func(ctx context.Context) error {
				time.Sleep(time.Second)
				return nil
			},
			err: context.DeadlineExceeded,
		},
	}

	for caseName, tc := range testcases {
		tc := tc

		t.Run(caseName, func(t *testing.T) {
			t.Parallel()

			err := synx.CallWithTimeout(tc.timeout, tc.fn)
			if diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("(-want +got):\\n%s", diff)
			}
		})
	}
}
