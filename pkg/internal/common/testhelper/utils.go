package testhelper

import (
	"context"
	"errors"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testResourceName      = "test-resource-name"
	testResourceNamespace = "test-resource-namespace"
	testAnnotationKey     = "test-annotation-key"
	testAnnotationValue   = "test-annotation-value"
)

var (
	errCreateFailure  = errors.New("simulated create failure")
	errDeleteFailure  = errors.New("simulated delete failure")
	errUpdateFailure  = errors.New("simulated update failure")
	errListFailure    = errors.New("simulated list failure")
	errInvalidBuilder = errors.New("invalid builder error")
)

var testFailingCreate = func(
	ctx context.Context,
	client runtimeclient.WithWatch,
	obj runtimeclient.Object,
	opts ...runtimeclient.CreateOption,
) error {
	return errCreateFailure
}

var testFailingDelete = func(
	ctx context.Context,
	client runtimeclient.WithWatch,
	obj runtimeclient.Object,
	opts ...runtimeclient.DeleteOption,
) error {
	return errDeleteFailure
}

var testFailingUpdate = func(
	ctx context.Context,
	client runtimeclient.WithWatch,
	obj runtimeclient.Object,
	opts ...runtimeclient.UpdateOption,
) error {
	return errUpdateFailure
}

var testFailingList = func(
	ctx context.Context,
	client runtimeclient.WithWatch,
	list runtimeclient.ObjectList,
	opts ...runtimeclient.ListOption,
) error {
	return errListFailure
}

func isErrorNil(err error) bool {
	return err == nil
}

func isAPICallFailedWithCreate(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "create")
}

func isAPICallFailedWithDelete(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "delete")
}

func isAPICallFailedWithUpdate(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "update")
}

func isAPICallFailedWithGet(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "get")
}

func isAPICallFailedWithList(err error) bool {
	return commonerrors.IsAPICallFailedWithVerb(err, "list")
}

func isInvalidBuilder(err error) bool {
	return errors.Is(err, errInvalidBuilder)
}

func isContextDeadlineExceeded(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

func buildDummyObject[O any, SO common.ObjectPointer[O]](name, namespace string) SO {
	var dummyObject SO = new(O)

	dummyObject.SetName(name)
	dummyObject.SetNamespace(namespace)

	return dummyObject
}

// createSchemeAttacherGVKTest creates a test function that checks if the scheme attacher registers the expected GVK.
// The returned function must be run with t.Run().
func createSchemeAttacherGVKTest[O any, SO common.ObjectPointer[O]](
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		t.Parallel()

		scheme := runtime.NewScheme()
		err := schemeAttacher(scheme)
		require.NoError(t, err, "schemeAttacher failed when attaching to a fresh scheme")

		var obj SO = new(O)

		kinds, _, err := scheme.ObjectKinds(obj)
		assert.NoError(t, err, "scheme.ObjectKinds failed for test object; scheme attacher may be wrong")
		assert.Contains(t, kinds, expectedGVK, "scheme attacher did not register the expected GVK")
	}
}
