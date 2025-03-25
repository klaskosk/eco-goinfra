package watch

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	toolswatch "k8s.io/client-go/tools/watch"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type objectListRef[T any] interface {
	runtimeclient.ObjectList
	*T
}

type runtimeWatcher[T any, PT objectListRef[T]] struct {
	runtimeclient.WithWatch
	runtimeclient.ObjectKey
}

// RetryFromRuntimeClient creates a RetryWatcher that selects only the provided object. The object's ResourceVersion is
// used to start the watcher so it should be as recent as possible to avoid receiving old events. Object is guaranteed
// not to be mutated.
func RetryFromRuntimeClient[T any, PT objectListRef[T]](
	apiClient runtimeclient.WithWatch, object runtimeclient.Object) (*toolswatch.RetryWatcher, error) {
	watcher := runtimeWatcher[T, PT]{
		WithWatch: apiClient,
		ObjectKey: runtimeclient.ObjectKeyFromObject(object),
	}

	return toolswatch.NewRetryWatcher(object.GetResourceVersion(), watcher)
}

// Watch implements cache.Watcher.
//
//nolint:ireturn // need to return an interface to satisfy cache.Watcher
func (watcher runtimeWatcher[T, PT]) Watch(options metav1.ListOptions) (watch.Interface, error) {
	var objectList PT = new(T)

	runtimeOptions := &runtimeclient.ListOptions{
		Namespace:     watcher.Namespace,
		FieldSelector: fields.OneTermEqualSelector(metav1.ObjectNameField, watcher.Name),
		Raw:           &options,
	}

	return watcher.WithWatch.Watch(context.TODO(), objectList, runtimeOptions)
}
