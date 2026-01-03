package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NamespacedPullFunc is a namespaced Pull function signature (e.g., PullHFS, PullHFC).
type NamespacedPullFunc[SB any] func(apiClient *clients.Settings, name, nsname string) (SB, error)

// ClusterScopedPullFunc is a cluster-scoped Pull function signature.
type ClusterScopedPullFunc[SB any] func(apiClient *clients.Settings, name string) (SB, error)

// PullTestConfig provides the configuration needed to test a Pull function wrapper.
type PullTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
] struct {
	SchemeAttacher clients.SchemeAttacher
	ExpectedGVK    schema.GroupVersionKind
	ResourceScope  ResourceScope
	pullFunc       NamespacedPullFunc[SB]
}

// NewNamespacedPullTestConfig creates a new PullTestConfig for namespaced resources.
func NewNamespacedPullTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
](
	pullFunc NamespacedPullFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) PullTestConfig[O, B, SO, SB] {
	return PullTestConfig[O, B, SO, SB]{
		SchemeAttacher: schemeAttacher,
		ExpectedGVK:    expectedGVK,
		ResourceScope:  ResourceScopeNamespaced,
		pullFunc:       pullFunc,
	}
}

// NewClusterScopedPullTestConfig creates a new PullTestConfig for cluster-scoped resources.
// The cluster-scoped pull function is wrapped in a closure that ignores the namespace parameter.
func NewClusterScopedPullTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
](
	pullFunc ClusterScopedPullFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) PullTestConfig[O, B, SO, SB] {
	return PullTestConfig[O, B, SO, SB]{
		SchemeAttacher: schemeAttacher,
		ExpectedGVK:    expectedGVK,
		ResourceScope:  ResourceScopeClusterScoped,
		pullFunc: func(apiClient *clients.Settings, name, _ string) (SB, error) {
			return pullFunc(apiClient, name)
		},
	}
}

// Name returns the name to use for running these tests.
func (config PullTestConfig[O, B, SO, SB]) Name() string {
	return "Pull"
}

// ExecuteTests runs the standard set of tests for a Pull function wrapper.
//
//nolint:funlen // Test function with multiple test cases.
func (config PullTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	type testCase struct {
		name          string
		clientNil     bool
		builderName   string
		builderNsName string
		objectExists  bool
		assertError   func(error) bool
	}

	testCases := []testCase{
		{
			name:          "valid pull existing resource",
			clientNil:     false,
			builderName:   testResourceName,
			builderNsName: testResourceNamespace,
			objectExists:  true,
			assertError:   isErrorNil,
		},
		{
			name:          "nil client returns error",
			clientNil:     true,
			builderName:   testResourceName,
			builderNsName: testResourceNamespace,
			objectExists:  false,
			assertError:   commonerrors.IsAPIClientNil,
		},
		{
			name:          "empty name returns error",
			clientNil:     false,
			builderName:   "",
			builderNsName: testResourceNamespace,
			objectExists:  false,
			assertError:   commonerrors.IsBuilderNameEmpty,
		},
		{
			name:          "non-existent resource returns not found",
			clientNil:     false,
			builderName:   "non-existent-resource",
			builderNsName: "non-existent-namespace",
			objectExists:  false,
			assertError:   k8serrors.IsNotFound,
		},
	}

	if config.ResourceScope.IsNamespaced() {
		testCases = append(testCases, testCase{
			name:          "empty namespace returns error",
			clientNil:     false,
			builderName:   testResourceName,
			builderNsName: "",
			objectExists:  false,
			assertError:   commonerrors.IsBuilderNamespaceEmpty,
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
				if testCase.objectExists {
					var namespace string
					if config.ResourceScope.IsNamespaced() {
						namespace = testCase.builderNsName
					}

					objects = append(objects, buildDummyObject[O, SO](testCase.builderName, namespace))
				}

				client = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects:  objects,
					SchemeAttachers: []clients.SchemeAttacher{config.SchemeAttacher},
				})
			}

			builder, err := config.pullFunc(client, testCase.builderName, testCase.builderNsName)

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				require.NotNil(t, builder)
				require.NotNil(t, builder.GetDefinition())

				assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testCase.builderNsName, builder.GetDefinition().GetNamespace())
				}

				require.NotNil(t, builder.GetObject())
				assert.Equal(t, testCase.builderName, builder.GetObject().GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testCase.builderNsName, builder.GetObject().GetNamespace())
				}

				assert.Equal(t, config.ExpectedGVK, builder.GetGVK())
			} else {
				assert.Nil(t, builder)
			}
		})
	}
}
