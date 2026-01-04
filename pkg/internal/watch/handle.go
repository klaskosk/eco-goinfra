package watch

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/watch"
	toolswatch "k8s.io/client-go/tools/watch"
)

// ConditionFunc is called on every add, update, or delete event. If an error is returned, the watch task will stop and
// Wait returns the same error. Otherwise, we loop until bool is true. The provided context is the same on every call.
type ConditionFunc func(ctx context.Context, event watch.Event) (bool, error)

// Handle provides access to the results of an asynchronous watch condition.
type Handle struct {
	watcher watch.Interface
	cond    ConditionFunc
	done    chan struct{}
	err     error
}

// NewHandle eagerly starts the watch goroutine and returns a handle to its results. Once the goroutine finishes, it
// will never be restarted and its result is saved in the handle.
func NewHandle(ctx context.Context, watcher watch.Interface, cond ConditionFunc) *Handle {
	return NewHandleWithTimeout(ctx, watcher, 0, cond)
}

// NewHandleWithTimeout works identically to NewHandle except the watch goroutine will finish if the timeout has passed.
func NewHandleWithTimeout(
	ctx context.Context, watcher watch.Interface, timeout time.Duration, cond ConditionFunc) *Handle {
	handle := &Handle{
		cond:    cond,
		watcher: watcher,
		done:    make(chan struct{}),
		err:     nil,
	}

	go handle.start(ctx, timeout)

	return handle
}

// Wait blocks until the handle has finished, returning nil on success and an error if one occurred. If the handle is
// finished, this returns the same result immediately.
func (handle *Handle) Wait() error {
	<-handle.done

	return handle.err
}

// start runs the main watching loop. It is blocking and should only ever be called once when a Handle is created.
func (handle *Handle) start(ctx context.Context, timeout time.Duration) {
	defer close(handle.done)
	defer handle.watcher.Stop()

	ctx, cancel := toolswatch.ContextWithOptionalTimeout(ctx, timeout)
	defer cancel()

	if handle.cond == nil {
		handle.err = fmt.Errorf("condition function is nil")

		return
	}

	for {
		select {
		case event, ok := <-handle.watcher.ResultChan():
			if !ok {
				handle.err = fmt.Errorf("watcher results closed unexpectedly")

				return
			}

			if event.Type == watch.Error {
				err := apierrors.FromObject(event.Object)
				handle.err = fmt.Errorf("watcher sent an error event: %w", err)

				return
			}

			// Skip bookmarks since they are not guaranteed to be returned and do not affect the condition.
			if event.Type == watch.Bookmark {
				continue
			}

			conditionMet, err := handle.cond(ctx, event)
			if err != nil {
				handle.err = fmt.Errorf("condition function returned non-nil error: %w", err)

				return
			}

			if conditionMet {
				return
			}
		case <-ctx.Done():
			handle.err = fmt.Errorf("handle context done: %w", ctx.Err())

			return
		}
	}
}
