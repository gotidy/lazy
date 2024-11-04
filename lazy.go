package lazy

import (
	"context"
	"iter"
	"sync"
	"time"

	"github.com/gotidy/iters"
)

type options struct {
	retryDelays iter.Seq[time.Duration]
}

type option func(opts *options)

func WithRetry(delays iter.Seq[time.Duration]) func(opts *options) {
	return func(opts *options) {
		opts.retryDelays = delays
	}
}

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
		for range iters.RetryAfterDelay(ctx, options.retryDelays) {
			retryObj, retryErr := creator(ctx)
			mu.Lock()
			obj, err = retryObj, retryErr
			mu.Unlock()
			if err == nil {
				break
			}
		}
	}()

	return func(ctx context.Context) (T, error) {
		select {
		case <-done:
		case <-ctx.Done():
			return obj, ctx.Err()
		}
		mu.Lock()
		defer mu.Unlock()
		return obj, err
	}
}
