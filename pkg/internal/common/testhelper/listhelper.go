package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// ListInAllNamespacesFunc is a List function signature for listing resources in all namespaces
// (e.g., ListInAllNamespaces).
type ListInAllNamespacesFunc[SB any] func(
	apiClient *clients.Settings,
	options ...runtimeclient.ListOptions,
) ([]SB, error)

// NamespacedListFunc is a List function signature for listing resources in a specific namespace
// (e.g., List with namespace parameter).
type NamespacedListFunc[SB any] func(
	apiClient *clients.Settings,
	nsname string,
	options ...runtimeclient.ListOptions,
) ([]SB, error)

// ListTestConfig provides the configuration needed to test a List function wrapper.
type ListTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
] struct {
	SchemeAttacher clients.SchemeAttacher
	ExpectedGVK    schema.GroupVersionKind
	ResourceScope  ResourceScope
	listFunc       NamespacedListFunc[SB]
}

// NewListTestConfig creates a new ListTestConfig for ListInAllNamespaces-style functions.
func NewListTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
](
	listFunc ListInAllNamespacesFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) ListTestConfig[O, B, SO, SB] {
	return ListTestConfig[O, B, SO, SB]{
		SchemeAttacher: schemeAttacher,
		ExpectedGVK:    expectedGVK,
		ResourceScope:  ResourceScopeClusterScoped,
		listFunc: func(apiClient *clients.Settings, _ string, options ...runtimeclient.ListOptions) ([]SB, error) {
			return listFunc(apiClient, options...)
		},
	}
}

// NewNamespacedListTestConfig creates a new ListTestConfig for namespaced List functions.
func NewNamespacedListTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
](
	listFunc NamespacedListFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) ListTestConfig[O, B, SO, SB] {
	return ListTestConfig[O, B, SO, SB]{
		SchemeAttacher: schemeAttacher,
		ExpectedGVK:    expectedGVK,
		ResourceScope:  ResourceScopeNamespaced,
		listFunc:       listFunc,
	}
}

// Name returns the name to use for running these tests.
func (config ListTestConfig[O, B, SO, SB]) Name() string {
	return "List"
}

// ExecuteTests runs the standard set of tests for a List function wrapper.
//
//nolint:funlen // Test function with multiple test cases.
func (config ListTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	type testCase struct {
		name             string
		clientNil        bool
		nsname           string
		objectsExist     bool
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
		expectedCount    int
	}

	testCases := []testCase{
		{
			name:          "valid list with resources",
			clientNil:     false,
			nsname:        testResourceNamespace,
			objectsExist:  true,
			assertError:   isErrorNil,
			expectedCount: 2,
		},
		{
			name:          "valid list empty",
			clientNil:     false,
			nsname:        testResourceNamespace,
			objectsExist:  false,
			assertError:   isErrorNil,
			expectedCount: 0,
		},
		{
			name:          "nil client returns error",
			clientNil:     true,
			nsname:        testResourceNamespace,
			objectsExist:  false,
			assertError:   commonerrors.IsAPIClientNil,
			expectedCount: 0,
		},
		{
			name:             "failed list call returns error",
			clientNil:        false,
			nsname:           testResourceNamespace,
			objectsExist:     false,
			interceptorFuncs: interceptor.Funcs{List: testFailingList},
			assertError:      isAPICallFailedWithList,
			expectedCount:    0,
		},
	}

	if config.ResourceScope.IsNamespaced() {
		testCases = append(testCases, testCase{
			name:          "empty namespace returns error",
			clientNil:     false,
			nsname:        "",
			objectsExist:  false,
			assertError:   commonerrors.IsBuilderNamespaceEmpty,
			expectedCount: 0,
		})
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				client  *clients.Settings
				objects []runtime.Object
			)

			if !testCase.clientNil {
				if testCase.objectsExist {
					objects = append(objects,
						buildDummyObject[O, SO]("resource-1", testResourceNamespace),
						buildDummyObject[O, SO]("resource-2", testResourceNamespace),
					)
				}

				client = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:   objects,
					SchemeAttachers:  []clients.SchemeAttacher{config.SchemeAttacher},
					InterceptorFuncs: testCase.interceptorFuncs,
				})
			}

			builders, err := config.listFunc(client, testCase.nsname)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				assert.Len(t, builders, testCase.expectedCount)

				for _, builder := range builders {
					require.NotNil(t, builder.GetDefinition())
					require.NotNil(t, builder.GetObject())
					require.NotNil(t, builder.GetClient())
				}
			} else {
				assert.Empty(t, builders)
			}
		})
	}
}
