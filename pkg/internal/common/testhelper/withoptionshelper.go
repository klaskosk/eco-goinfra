package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WithOptionsUser is an interface for builders that have a WithOptions method. AO is the option type (typically a
// named func type such as configmap.AdditionalOptions, or the plain func(SB) (SB, error) form); it must match
// common.AdditionalOption[SB].
type WithOptionsUser[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
	AO common.AdditionalOption[SB],
] interface {
	common.BuilderPointer[B, O, SO]
	WithOptions(options ...AO) SB
}

// WithOptionsTestConfig provides the configuration needed to test a WithOptions method.
type WithOptionsTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB common.BuilderPointer[B, O, SO],
	AO common.AdditionalOption[SB],
] struct {
	CommonTestConfig[O, B, SO, SB]

	withOptionsFunc func(builder SB, options ...AO) SB
}

// NewWithOptionsTestConfig creates a new WithOptionsTestConfig with the given parameters for builders that implement
// the WithOptionsUser interface.
func NewWithOptionsTestConfig[
	O, B any,
	SO common.ObjectPointer[O],
	SB WithOptionsUser[O, B, SO, SB, AO],
	AO common.AdditionalOption[SB],
](commonTestConfig CommonTestConfig[O, B, SO, SB]) WithOptionsTestConfig[O, B, SO, SB, AO] {
	return WithOptionsTestConfig[O, B, SO, SB, AO]{
		CommonTestConfig: commonTestConfig,
		withOptionsFunc: func(builder SB, options ...AO) SB {
			return builder.WithOptions(options...)
		},
	}
}

// Name returns the name to use for running these tests.
func (config WithOptionsTestConfig[O, B, SO, SB, AO]) Name() string {
	return "WithOptions"
}

// ExecuteTests runs the standard set of WithOptions tests for the configured resource.
func (config WithOptionsTestConfig[O, B, SO, SB, AO]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	testCases := []struct {
		name         string
		builderError error
		options      []AO
		assertError  func(error) bool
		expectedName string
	}{
		{
			name:        "options are applied successfully",
			options:     []AO{testAnnotationOption[O, B, SO, SB, AO]()},
			assertError: isErrorNil,
		},
		{
			name:         "invalid builder stops processing",
			builderError: errInvalidBuilder,
			options:      []AO{testAnnotationOption[O, B, SO, SB, AO]()},
			assertError:  isInvalidBuilder,
		},
		{
			name:        "error in option is captured in builder",
			options:     []AO{testFailingOption[O, B, SO, SB, AO]()},
			assertError: isOptionFailure,
		},
		{
			name:        "nil options are skipped",
			options:     []AO{nil, testAnnotationOption[O, B, SO, SB, AO](), nil},
			assertError: isErrorNil,
		},
		{
			name:        "error stops subsequent options",
			options:     []AO{testFailingOption[O, B, SO, SB, AO](), testAnnotationOption[O, B, SO, SB, AO]()},
			assertError: isOptionFailure,
		},
		{
			name:        "nil builder from option does not clobber builder",
			options:     []AO{testNilBuilderOption[O, B, SO, SB, AO](), testAnnotationOption[O, B, SO, SB, AO]()},
			assertError: isErrorNil,
		},
		{
			name:         "error with replacement builder uses replacement",
			options:      []AO{testFailingOptionWithReplacementBuilder[O, B, SO, SB, AO]()},
			assertError:  isOptionFailure,
			expectedName: "replacement",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := config.buildTestBuilder(t)
			builder.SetError(testCase.builderError)

			result := config.withOptionsFunc(builder, testCase.options...)

			config.assertResult(t, result, testCase.assertError, testCase.expectedName)
		})
	}
}

// buildTestBuilder creates a valid builder for testing, scoped appropriately for the resource type.
func (config WithOptionsTestConfig[O, B, SO, SB, AO]) buildTestBuilder(t *testing.T) SB {
	t.Helper()

	client := clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{config.SchemeAttacher},
	})

	if config.ResourceScope.IsNamespaced() {
		return common.NewNamespacedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName, testResourceNamespace)
	}

	return common.NewClusterScopedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName)
}

// assertResult verifies the result of a WithOptions call: correct error, expected resource name, and annotation state
// consistent with whether the option succeeded or failed.
func (config WithOptionsTestConfig[O, B, SO, SB, AO]) assertResult(
	t *testing.T, result SB, assertError func(error) bool, expectedName string,
) {
	t.Helper()

	require.NotNil(t, result)
	require.Truef(t, assertError(result.GetError()), "unexpected error, got: %v", result.GetError())

	if expectedName == "" {
		expectedName = testResourceName
	}

	assert.Equal(t, expectedName, result.GetDefinition().GetName(),
		"returned builder should have the expected resource name")

	if result.GetError() == nil {
		annotations := result.GetDefinition().GetAnnotations()
		require.NotNil(t, annotations)
		assert.Equal(t, testAnnotationValue, annotations[testAnnotationKey])
	} else {
		annotations := result.GetDefinition().GetAnnotations()
		if annotations != nil {
			_, exists := annotations[testAnnotationKey]
			assert.False(t, exists, "annotation should not be set when option fails")
		}
	}
}
