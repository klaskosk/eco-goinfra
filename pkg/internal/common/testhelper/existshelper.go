package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

// Exister is an interface for builders that have an Exists method.
type Exister[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	Exists() bool
}

// ExistsTestConfig provides the configuration needed to test an Exists method.
type ExistsTestConfig[O, B any, SO common.ObjectPointer[O], SB Exister[O, B, SO, SB]] struct {
	CommonTestConfig[O, B, SO, SB]
}

// NewExistsTestConfig creates a new ExistsTestConfig with the given parameters.
func NewExistsTestConfig[O, B any, SO common.ObjectPointer[O], SB Exister[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) ExistsTestConfig[O, B, SO, SB] {
	return ExistsTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
	}
}

// Name returns the name to use for running these tests.
func (config ExistsTestConfig[O, B, SO, SB]) Name() string {
	return "Exists"
}

// ExecuteTests runs the standard set of Exists tests for the configured resource.
func (config ExistsTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	testCases := []struct {
		name           string
		objectExists   bool
		builderError   error
		expectedResult bool
	}{
		{
			name:           "valid exists returns true when resource exists",
			objectExists:   true,
			expectedResult: true,
		},
		{
			name:           "invalid builder returns false",
			objectExists:   true,
			builderError:   errInvalidBuilder,
			expectedResult: false,
		},
		{
			name:           "resource does not exist returns false",
			objectExists:   false,
			expectedResult: false,
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

			result := builder.Exists()

			assert.Equal(t, testCase.expectedResult, result)

			if testCase.expectedResult {
				require.NotNil(t, builder.GetObject())
				assert.Equal(t, testResourceName, builder.GetObject().GetName())

				if config.ResourceScope.IsNamespaced() {
					assert.Equal(t, testResourceNamespace, builder.GetObject().GetNamespace())
				}
			}
		})
	}
}
