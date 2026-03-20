//nolint:ireturn // Seems to be a bug in the linter that this needs to be at the package level.
package watch

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	errConditionFailed = errors.New("condition function error")
)

//nolint:funlen // This function is long due only due to the number of test cases.
func TestNewHandle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		setupFunc   func(*mockWatcher)
		cond        ConditionFunc
		assertError func(error) bool
	}{
		{
			name: "condition met on first event",
			setupFunc: func(w *mockWatcher) {
				w.sendEvent(watch.Event{Type: watch.Added, Object: &mockStatus{}})
			},
			cond: func(_ context.Context, _ watch.Event) (bool, error) {
				return true, nil
			},
			assertError: isErrorNil,
		},
		{
			name:      "nil condition function",
			setupFunc: func(w *mockWatcher) {},
			cond:      nil,
			assertError: func(err error) bool {
				return err != nil && err.Error() == "condition function is nil"
			},
		},
		{
			name: "watcher results closed unexpectedly",
			setupFunc: func(w *mockWatcher) {
				w.close()
			},
			cond: func(_ context.Context, _ watch.Event) (bool, error) {
				return false, nil
			},
			assertError: func(err error) bool {
				return err != nil && err.Error() == "watcher results closed unexpectedly"
			},
		},
		{
			name: "condition function returns error",
			setupFunc: func(w *mockWatcher) {
				w.sendEvent(watch.Event{Type: watch.Added, Object: &mockStatus{}})
			},
			cond: func(_ context.Context, _ watch.Event) (bool, error) {
				return false, errConditionFailed
			},
			assertError: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "condition function returned non-nil error")
			},
		},
		{
			name: "watcher sends error event",
			setupFunc: func(w *mockWatcher) {
				w.sendEvent(watch.Event{
					Type: watch.Error,
					Object: &mockStatus{
						Status: metav1.Status{
							Status:  metav1.StatusFailure,
							Message: "test error",
							Reason:  metav1.StatusReasonInternalError,
							Code:    500,
						},
					},
				})
			},
			cond: func(_ context.Context, _ watch.Event) (bool, error) {
				return false, nil
			},
			assertError: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "watcher sent an error event")
			},
		},
		{
			name: "bookmark events are skipped",
			setupFunc: func(w *mockWatcher) {
				// Send bookmark first (should be skipped)
				w.sendEvent(watch.Event{Type: watch.Bookmark, Object: &mockStatus{}})
				// Then send a real event that meets the condition
				w.sendEvent(watch.Event{Type: watch.Added, Object: &mockStatus{}})
			},
			cond: func(_ context.Context, _ watch.Event) (bool, error) {
				return true, nil
			},
			assertError: isErrorNil,
		},
		{
			name: "condition not met continues watching",
			setupFunc: func(w *mockWatcher) {
				// First event doesn't meet condition
				w.sendEvent(watch.Event{Type: watch.Added, Object: &mockStatus{}})
				// Second event meets condition
				w.sendEvent(watch.Event{Type: watch.Modified, Object: &mockStatus{}})
			},
			cond: func() ConditionFunc {
				callCount := 0

				return func(_ context.Context, _ watch.Event) (bool, error) {
					callCount++
					// Return true on second call
					return callCount >= 2, nil
				}
			}(),
			assertError: isErrorNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			watcher := newMockWatcher()

			handle := NewHandle(t.Context(), watcher, testCase.cond)
			testCase.setupFunc(watcher)

			err := handle.Wait()

			assert.Truef(t, testCase.assertError(err), "got error %v", err)
			assert.True(t, watcher.stopped, "watcher should be stopped")
		})
	}
}

func TestNewHandleWithTimeout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		timeout     time.Duration
		setupFunc   func(*mockWatcher)
		cond        ConditionFunc
		assertError func(error) bool
	}{
		{
			name:      "context timeout",
			timeout:   50 * time.Millisecond,
			setupFunc: func(_ *mockWatcher) {},
			cond: func(_ context.Context, _ watch.Event) (bool, error) {
				return false, nil
			},
			assertError: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "handle context done")
			},
		},
		{
			name:    "condition met before timeout",
			timeout: time.Second,
			setupFunc: func(w *mockWatcher) {
				w.sendEvent(watch.Event{Type: watch.Added, Object: &mockStatus{}})
			},
			cond: func(_ context.Context, _ watch.Event) (bool, error) {
				return true, nil
			},
			assertError: isErrorNil,
		},
		{
			name:    "zero timeout means no timeout",
			timeout: 0,
			setupFunc: func(w *mockWatcher) {
				w.sendEvent(watch.Event{Type: watch.Added, Object: &mockStatus{}})
			},
			cond: func(_ context.Context, _ watch.Event) (bool, error) {
				return true, nil
			},
			assertError: isErrorNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			watcher := newMockWatcher()

			handle := NewHandleWithTimeout(t.Context(), watcher, testCase.timeout, testCase.cond)

			testCase.setupFunc(watcher)

			err := handle.Wait()

			assert.Truef(t, testCase.assertError(err), "got error %v", err)
			assert.True(t, watcher.stopped, "watcher should be stopped")
		})
	}
}

// mockWatcher implements watch.Interface for testing.
type mockWatcher struct {
	resultChan chan watch.Event
	stopped    bool
}

func newMockWatcher() *mockWatcher {
	return &mockWatcher{
		resultChan: make(chan watch.Event, 10),
		stopped:    false,
	}
}

func (m *mockWatcher) Stop() {
	m.stopped = true
}

func (m *mockWatcher) ResultChan() <-chan watch.Event {
	return m.resultChan
}

// sendEvent sends an event to the mock watcher.
func (m *mockWatcher) sendEvent(event watch.Event) {
	m.resultChan <- event
}

// close closes the result channel.
func (m *mockWatcher) close() {
	close(m.resultChan)
}

// mockStatus implements the runtime.Object and metav1.Status for error events.
type mockStatus struct {
	metav1.Status
}

var _ runtime.Object = (*mockStatus)(nil)

func (m *mockStatus) GetObjectKind() schema.ObjectKind {
	return &metav1.TypeMeta{Kind: "Status", APIVersion: "v1"}
}

func (m *mockStatus) DeepCopyObject() runtime.Object {
	return &mockStatus{Status: m.Status}
}

// isErrorNil returns true if the error is nil.
func isErrorNil(err error) bool {
	return err == nil
}
