package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// NamespacedNewBuilderFunc is a namespaced NewBuilder function signature.
type NamespacedNewBuilderFunc[SB any] func(apiClient *clients.Settings, name, nsname string) SB

// ClusterScopedNewBuilderFunc is a cluster-scoped NewBuilder function signature.
type ClusterScopedNewBuilderFunc[SB any] func(apiClient *clients.Settings, name string) SB

// NewBuilderTestConfig provides the configuration needed to test a NewBuilder function wrapper.
type NewBuilderTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
] struct {
	SchemeAttacher clients.SchemeAttacher
	ExpectedGVK    schema.GroupVersionKind
	ResourceScope  ResourceScope
	newBuilderFunc NamespacedNewBuilderFunc[SB]
}

// NewNamespacedNewBuilderTestConfig creates a new NewBuilderTestConfig for namespaced resources.
func NewNamespacedNewBuilderTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
](
	newBuilderFunc NamespacedNewBuilderFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) NewBuilderTestConfig[O, B, SO, SB] {
	return NewBuilderTestConfig[O, B, SO, SB]{
		SchemeAttacher: schemeAttacher,
		ExpectedGVK:    expectedGVK,
		ResourceScope:  ResourceScopeNamespaced,
		newBuilderFunc: newBuilderFunc,
	}
}

// NewClusterScopedNewBuilderTestConfig creates a new NewBuilderTestConfig for cluster-scoped resources. The
// cluster-scoped NewBuilder function is wrapped in a closure that ignores the namespace parameter.
func NewClusterScopedNewBuilderTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
](
	newBuilderFunc ClusterScopedNewBuilderFunc[SB],
	schemeAttacher clients.SchemeAttacher,
	expectedGVK schema.GroupVersionKind,
) NewBuilderTestConfig[O, B, SO, SB] {
	return NewBuilderTestConfig[O, B, SO, SB]{
		SchemeAttacher: schemeAttacher,
		ExpectedGVK:    expectedGVK,
		ResourceScope:  ResourceScopeClusterScoped,
		newBuilderFunc: func(apiClient *clients.Settings, name, _ string) SB {
			return newBuilderFunc(apiClient, name)
		},
	}
}

// Name returns the name to use for running these tests.
func (config NewBuilderTestConfig[O, B, SO, SB]) Name() string {
	return "NewBuilder"
}

// ExecuteTests runs the standard set of tests for a NewBuilder function.
func (config NewBuilderTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	type testCase struct {
		name          string
		clientNil     bool
		builderName   string
		builderNsName string
		assertError   func(error) bool
	}

	testCases := []testCase{
		{
			name:          "valid builder creation",
			clientNil:     false,
			builderName:   testResourceName,
			builderNsName: testResourceNamespace,
			assertError:   isErrorNil,
		},
		{
			name:          "nil client returns error",
			clientNil:     true,
			builderName:   testResourceName,
			builderNsName: testResourceNamespace,
			assertError:   commonerrors.IsAPIClientNil,
		},
		{
			name:          "empty name returns error",
			clientNil:     false,
			builderName:   "",
			builderNsName: testResourceNamespace,
			assertError:   commonerrors.IsBuilderNameEmpty,
		},
	}

	if config.ResourceScope.IsNamespaced() {
		testCases = append(testCases, testCase{
			name:          "empty namespace returns error",
			clientNil:     false,
			builderName:   testResourceName,
			builderNsName: "",
			assertError:   commonerrors.IsBuilderNamespaceEmpty,
		})
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var client *clients.Settings

			if !testCase.clientNil {
				client = clients.GetTestClients(clients.TestClientParams{
					SchemeAttachers: []clients.SchemeAttacher{config.SchemeAttacher},
				})
			}

			builder := config.newBuilderFunc(client, testCase.builderName, testCase.builderNsName)

			require.NotNil(t, builder)
			require.Truef(t, testCase.assertError(builder.GetError()), "unexpected error, got: %v", builder.GetError())

			if builder.GetError() == nil {
				require.NotNil(t, builder.GetDefinition())
				assert.Equal(t, testCase.builderName, builder.GetDefinition().GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testCase.builderNsName, builder.GetDefinition().GetNamespace())
				}

				require.NotNil(t, builder.GetClient())
				assert.Implements(t, (*runtimeclient.Client)(nil), builder.GetClient())

				assert.Equal(t, config.ExpectedGVK, builder.GetGVK())
			}
		})
	}
}
