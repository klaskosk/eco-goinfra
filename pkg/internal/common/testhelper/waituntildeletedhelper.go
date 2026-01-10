package testhelper

import (
	"context"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// testTimeout is a short timeout used for testing to speed up test execution.
const testTimeout = 50 * time.Millisecond

// WaitUntilDeleter is an interface for builders that have a WaitUntilDeleted method.
type WaitUntilDeleter[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	WaitUntilDeleted(timeout time.Duration) error
}

// internalWaitUntilDeletedFunc defines the signature for wait until deleted operations.
type internalWaitUntilDeletedFunc[
	O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO],
] func(ctx context.Context, builder SB, timeout time.Duration) error

// GenericWaitUntilDeletedFunc is the signature for the common.WaitUntilDeleted function.
type GenericWaitUntilDeletedFunc[O any, SO common.ObjectPointer[O]] func(
	ctx context.Context, builder common.Builder[O, SO], timeout time.Duration) error

// WaitUntilDeletedTestConfig provides the configuration needed to test a WaitUntilDeleted method.
type WaitUntilDeletedTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// waitUntilDeletedFunc is a function that waits until the resource is deleted and returns an error.
	waitUntilDeletedFunc internalWaitUntilDeletedFunc[O, B, SO, SB]
}

// NewWaitUntilDeletedTestConfig creates a new WaitUntilDeletedTestConfig with the given parameters for builders that
// implement the WaitUntilDeleter interface.
func NewWaitUntilDeletedTestConfig[O, B any, SO common.ObjectPointer[O], SB WaitUntilDeleter[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) WaitUntilDeletedTestConfig[O, B, SO, SB] {
	return WaitUntilDeletedTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		waitUntilDeletedFunc: func(_ context.Context, builder SB, timeout time.Duration) error {
			return builder.WaitUntilDeleted(timeout)
		},
	}
}

// NewGenericWaitUntilDeletedTestConfig creates a new WaitUntilDeletedTestConfig with a custom wait function. This is
// useful for testing standalone functions like common.WaitUntilDeleted() rather than builder methods.
func NewGenericWaitUntilDeletedTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	waitUntilDeletedFunc GenericWaitUntilDeletedFunc[O, SO],
) WaitUntilDeletedTestConfig[O, B, SO, SB] {
	return WaitUntilDeletedTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		waitUntilDeletedFunc: func(ctx context.Context, builder SB, timeout time.Duration) error {
			return waitUntilDeletedFunc(ctx, builder, timeout)
		},
	}
}

// Name returns the name to use for running these tests.
func (config WaitUntilDeletedTestConfig[O, B, SO, SB]) Name() string {
	return "WaitUntilDeleted"
}

// ExecuteTests runs the standard set of WaitUntilDeleted tests for the configured resource.
func (config WaitUntilDeletedTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
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
			name:         "resource already deleted succeeds",
			objectExists: false,
			assertError:  isErrorNil,
		},
		{
			name:         "invalid builder returns error",
			objectExists: false,
			builderError: errInvalidBuilder,
			assertError:  isInvalidBuilder,
		},
		{
			name:         "timeout waiting for deletion",
			objectExists: true,
			assertError:  isContextDeadlineExceeded,
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

			err := config.waitUntilDeletedFunc(t.Context(), builder, testTimeout)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)
		})
	}
}
