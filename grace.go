package synx

import (
	"context"
	"fmt"
)

// Graceful runs given function with given context and catch panics, which return as error.
func Graceful(ctx context.Context, fn func(context.Context) error) (err error) {
	defer func() {
		if errRecovery := recover(); errRecovery != nil {
			err = fmt.Errorf("%v", errRecovery)
		}
	}()

	return fn(ctx)
}
