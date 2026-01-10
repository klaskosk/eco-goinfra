package testhelper

import (
	"context"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// DeleteAndWaiter is an interface for builders that have a DeleteAndWait method returning only an error.
type DeleteAndWaiter[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	DeleteAndWait(timeout time.Duration) error
}

// DeleteAndWaitReturner is an interface for builders that have a DeleteAndWait method returning the builder and an
// error.
type DeleteAndWaitReturner[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	DeleteAndWait(timeout time.Duration) (SB, error)
}

// internalDeleteAndWaitFunc is the internal function signature used by DeleteAndWaitTestConfig. All of the other
// delete and wait functions must be able to be wrapped in this signature.
type internalDeleteAndWaitFunc[
	O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO],
] func(ctx context.Context, builder SB, timeout time.Duration) error

// GenericDeleteAndWaitFunc is the signature for the common.DeleteAndWait function.
type GenericDeleteAndWaitFunc[O any, SO common.ObjectPointer[O]] func(
	ctx context.Context, builder common.Builder[O, SO], timeout time.Duration) error

// DeleteAndWaitTestConfig provides the configuration needed to test a DeleteAndWait method.
type DeleteAndWaitTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// deleteAndWaitFunc is a function that deletes the resource and waits, returning an error. It gets set by the
	// constructor methods and will handle the different signatures of the DeleteAndWait method.
	deleteAndWaitFunc internalDeleteAndWaitFunc[O, B, SO, SB]
}

// NewDeleteAndWaiterTestConfig creates a new DeleteAndWaitTestConfig for builders that implement the DeleteAndWaiter
// interface.
func NewDeleteAndWaiterTestConfig[O, B any, SO common.ObjectPointer[O], SB DeleteAndWaiter[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) DeleteAndWaitTestConfig[O, B, SO, SB] {
	return DeleteAndWaitTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		deleteAndWaitFunc: func(_ context.Context, builder SB, timeout time.Duration) error {
			return builder.DeleteAndWait(timeout)
		},
	}
}

// NewDeleteAndWaitReturnerTestConfig creates a new DeleteAndWaitTestConfig for builders that implement the
// DeleteAndWaitReturner interface.
func NewDeleteAndWaitReturnerTestConfig[O, B any, SO common.ObjectPointer[O], SB DeleteAndWaitReturner[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) DeleteAndWaitTestConfig[O, B, SO, SB] {
	return DeleteAndWaitTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		deleteAndWaitFunc: func(_ context.Context, builder SB, timeout time.Duration) error {
			_, err := builder.DeleteAndWait(timeout)

			return err
		},
	}
}

// NewGenericDeleteAndWaitTestConfig creates a new DeleteAndWaitTestConfig with a custom delete and wait function. This
// is useful for testing standalone functions like common.DeleteAndWait() rather than builder methods.
func NewGenericDeleteAndWaitTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	deleteAndWaitFunc GenericDeleteAndWaitFunc[O, SO],
) DeleteAndWaitTestConfig[O, B, SO, SB] {
	return DeleteAndWaitTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		deleteAndWaitFunc: func(ctx context.Context, builder SB, timeout time.Duration) error {
			return deleteAndWaitFunc(ctx, builder, timeout)
		},
	}
}

// Name returns the name to use for running these tests.
func (config DeleteAndWaitTestConfig[O, B, SO, SB]) Name() string {
	return "DeleteAndWait"
}

// ExecuteTests runs the standard set of DeleteAndWait tests for the configured resource.
func (config DeleteAndWaitTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	testCases := []struct {
		name             string
		objectExists     bool
		builderError     error
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
	}{
		{
			name:         "valid delete and wait",
			objectExists: true,
			assertError:  isErrorNil,
		},
		{
			name:         "invalid builder returns error",
			objectExists: true,
			builderError: errInvalidBuilder,
			assertError:  isInvalidBuilder,
		},
		{
			name:         "resource does not exist succeeds",
			objectExists: false,
			assertError:  isErrorNil,
		},
		{
			name:             "failed deletion returns error",
			objectExists:     true,
			interceptorFuncs: interceptor.Funcs{Delete: testFailingDelete},
			assertError:      isAPICallFailedWithDelete,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object

			if testCase.objectExists {
				var namespace string
				if config.ResourceScope.IsNamespaced() {
					namespace = testResourceNamespace
				}

				objects = append(objects, buildDummyObject[O, SO](testResourceName, namespace))
			}

			client := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:   objects,
				SchemeAttachers:  []clients.SchemeAttacher{config.SchemeAttacher},
				InterceptorFuncs: testCase.interceptorFuncs,
			})

			var builder SB
			if config.ResourceScope.IsNamespaced() {
				builder = common.NewNamespacedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName, testResourceNamespace)
			} else {
				builder = common.NewClusterScopedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName)
			}

			builder.SetError(testCase.builderError)

			err := config.deleteAndWaitFunc(t.Context(), builder, testTimeout)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				assert.Nil(t, builder.GetObject())
			}
		})
	}
}
