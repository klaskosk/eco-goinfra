package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

// Getter is an interface for builders that have a Get method.
type Getter[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Get() (SO, error)
}

// GetTestConfig provides the configuration needed to test a Get method.
type GetTestConfig[O, B any, SO common.ObjectPointer[O], SB Getter[O, B, SO, SB]] struct {
	CommonTestConfig[O, B, SO, SB]
}

// NewGetTestConfig creates a new GetTestConfig with the given parameters.
func NewGetTestConfig[O, B any, SO common.ObjectPointer[O], SB Getter[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) GetTestConfig[O, B, SO, SB] {
	return GetTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
	}
}

// Name returns the name to use for running these tests.
func (config GetTestConfig[O, B, SO, SB]) Name() string {
	return "Get"
}

// ExecuteTests runs the standard set of Get tests for the configured resource.
func (config GetTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	testCases := []struct {
		name         string
		objectExists bool
		builderError error
		assertError  func(error) bool
	}{
		{
			name:         "valid get existing resource",
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
			name:         "resource does not exist returns not found",
			objectExists: false,
			assertError:  k8serrors.IsNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var objects []runtime.Object

			if testCase.objectExists {
				objects = append(objects, buildDummyObject[O, SO](testResourceName, testResourceNamespace))
			}

			client := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  objects,
				SchemeAttachers: []clients.SchemeAttacher{config.SchemeAttacher},
			})

			var builder SB
			if config.ResourceScope.IsNamespaced() {
				builder = common.NewNamespacedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName, testResourceNamespace)
			} else {
				builder = common.NewClusterScopedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName)
			}

			builder.SetError(testCase.builderError)

			result, err := builder.Get()

			require.Truef(t, testCase.assertError(err), "unexpected error, got: %v", err)

			if err == nil {
				require.NotNil(t, result)
				assert.Equal(t, testResourceName, result.GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testResourceNamespace, result.GetNamespace())
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
