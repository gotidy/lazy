// Package allows to initialize something in parallel without stopping the main process.
package lazy

import (
	"context"
	"fmt"
	"iter"
	"sync"
	"time"

	"github.com/gotidy/iters"
)

type options struct {
	retryDelays iter.Seq[time.Duration]
}

type option func(opts *options)

// WithRetry adds retries.
func WithRetry(delays iter.Seq[time.Duration]) func(opts *options) {
	return func(opts *options) {
		opts.retryDelays = delays
	}
}

// Me creates lazy initializer. It immediately returns a function that returns an object and triggers the creation of the object.
// After the first unsuccessful creator call, it returns the error, however will try to recreate the object, if retries are defined.
func Me[T any](ctx context.Context, creator func(ctx context.Context) (T, error), opts ...option) func(ctx context.Context) (T, error) {
	var obj T
	var err error
	var mu sync.Mutex

	var options options
	for _, opt := range opts {
		opt(&options)
	}

	done := make(chan struct{})
	go func() {
		obj, err = creator(ctx)
		close(done)
		if err == nil {
			return
		}
		for attempt, delay := range iters.RetryAfterDelay(ctx, options.retryDelays) {
			retryObj, retryErr := creator(ctx)
			if retryErr != nil {
				retryErr = fmt.Errorf("retry: attempt: %d; delay: %s: %v", attempt, delay, retryErr)
			}
			mu.Lock()
			obj, err = retryObj, retryErr
			mu.Unlock()
			if retryErr == nil {
				break
			}
		}
	}()

	return func(ctx context.Context) (T, error) {
		select {
		case <-done:
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		}
		mu.Lock()
		defer mu.Unlock()
		return obj, err
	}
}
