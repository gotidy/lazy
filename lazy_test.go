package lazy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gotidy/iters"
	"github.com/stretchr/testify/require"
)

func TestLazy(t *testing.T) {
	t.Parallel()

	t.Run("succeed", func(t *testing.T) {
		t.Parallel()

		fn := func(_ context.Context) (string, error) {
			time.Sleep(time.Millisecond * 100)
			return "hello", nil
		}

		ctx := context.Background()
		lazy := Me(ctx, fn)
		for range 5 {
			v, err := lazy(ctx)
			require.NoError(t, err, "first call should succeed")
			require.Equal(t, "hello", v, "first call should succeed")
			time.Sleep(time.Millisecond * 100)
		}
	})

	t.Run("first fail, context canceled", func(t *testing.T) {
		t.Parallel()

		i := 0
		done := make(chan struct{})
		fn := func(_ context.Context) (string, error) {
			if i > 0 {
				close(done)
				return "hello", nil
			}
			i++
			time.Sleep(time.Millisecond * 100)
			return "", errors.New("💩")
		}

		ctx := context.Background()
		lazy := Me(ctx, fn, WithRetry(iters.Of(time.Millisecond)))
		v, err := lazy(ctx)
		require.Error(t, err, "first call should fail")
		require.Equal(t, "", v, "first call result should be empty")
		<-done
		time.Sleep(time.Millisecond * 100)
		for range 5 {
			v, err := lazy(ctx)
			require.NoError(t, err, "next calls should succeed")
			require.Equal(t, "hello", v, "next calls result should be not empty")
			time.Sleep(time.Millisecond * 100)
		}
	})

	t.Run("first fail, context canceled", func(t *testing.T) {
		t.Parallel()

		done := make(chan struct{})
		fn := func(_ context.Context) (string, error) {
			<-done
			return "", errors.New("💩")
		}

		ctx, cancel := context.WithCancel(context.Background())
		lazy := Me(ctx, fn, WithRetry(iters.Of(time.Millisecond)))
		time.Sleep(time.Millisecond * 100)
		cancel()
		v, err := lazy(ctx)
		require.Error(t, err, "context canceled")
		require.Error(t, ctx.Err(), "context canceled")
		require.Equal(t, "", v, "context canceled")
		close(done)
	})
}
