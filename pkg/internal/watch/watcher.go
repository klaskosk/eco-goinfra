package watch

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	toolswatch "k8s.io/client-go/tools/watch"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPointer is a type constraint that requires a type be a pointer to L and implement the runtimeclient.ObjectList
// interface. This is duplicated from the common package to avoid import cycles.
type ListPointer[L any] interface {
	*L
	runtimeclient.ObjectList
}

type runtimeWatcher[L any, SL ListPointer[L]] struct {
	runtimeclient.WithWatch
	runtimeclient.ObjectKey
}

// NewRetryWatcherFromClient creates a RetryWatcher that selects only the provided object. The object's ResourceVersion
// is used to start the watcher so it should be as recent as possible to avoid receiving old events. Object is
// guaranteed not to be mutated by this function, but must have the resource version, name, and namespace (if
// namespaced) set.
//
// The list in the type parameters must correspond to the object, although this cannot be enforced at compile time.
func NewRetryWatcherFromClient[L any, SL ListPointer[L]](
	ctx context.Context, apiClient runtimeclient.WithWatch, object runtimeclient.Object) (*toolswatch.RetryWatcher, error) {
	watcher := runtimeWatcher[L, SL]{
		WithWatch: apiClient,
		ObjectKey: runtimeclient.ObjectKeyFromObject(object),
	}

	return toolswatch.NewRetryWatcherWithContext(ctx, object.GetResourceVersion(), watcher)
}

// Watch implements cache.WatcherWithContext.
//
//nolint:ireturn // need to return an interface to satisfy cache.WatcherWithContext
func (watcher runtimeWatcher[L, SL]) WatchWithContext(ctx context.Context, options metav1.ListOptions) (watch.Interface, error) {
	var objectList SL = new(L)

	metav1Options := &runtimeclient.ListOptions{Raw: &options}
	// For cluster-scoped objects, watcher.Namespace will be empty and this option will be ignored.
	namespaceOption := runtimeclient.InNamespace(watcher.Namespace)
	objectNameOption := runtimeclient.MatchingFields{metav1.ObjectNameField: watcher.Name}

	return watcher.Watch(ctx, objectList, namespaceOption, objectNameOption, metav1Options)
}
