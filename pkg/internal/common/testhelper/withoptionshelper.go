package testhelper

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WithOptionsUser is an interface for builders that have a WithOptions method.
type WithOptionsUser[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] interface {
	common.BuilderPointer[B, O, SO]
	WithOptions(options ...func(SB) (SB, error)) SB
}

// WithOptionsFunc defines the signature for with options operations.
type WithOptionsFunc[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] func(
	builder SB, options ...func(SB) (SB, error)) SB

// WithOptionsTestConfig provides the configuration needed to test a WithOptions method.
type WithOptionsTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]] struct {
	CommonTestConfig[O, B, SO, SB]

	// withOptionsFunc is a function that applies options to the builder and returns the result.
	withOptionsFunc WithOptionsFunc[O, B, SO, SB]
}

// NewWithOptionsTestConfig creates a new WithOptionsTestConfig with the given parameters for builders that implement
// the WithOptionsUser interface.
func NewWithOptionsTestConfig[O, B any, SO common.ObjectPointer[O], SB WithOptionsUser[O, B, SO, SB]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
) WithOptionsTestConfig[O, B, SO, SB] {
	return WithOptionsTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		withOptionsFunc: func(builder SB, options ...func(SB) (SB, error)) SB {
			return builder.WithOptions(options...)
		},
	}
}

// NewGenericWithOptionsTestConfig creates a new WithOptionsTestConfig with a custom with options function. This is
// useful for testing standalone functions like common.WithOptions() rather than builder methods.
func NewGenericWithOptionsTestConfig[O, B any, SO common.ObjectPointer[O], SB common.BuilderPointer[B, O, SO]](
	commonTestConfig CommonTestConfig[O, B, SO, SB],
	withOptionsFunc WithOptionsFunc[O, B, SO, SB],
) WithOptionsTestConfig[O, B, SO, SB] {
	return WithOptionsTestConfig[O, B, SO, SB]{
		CommonTestConfig: commonTestConfig,
		withOptionsFunc:  withOptionsFunc,
	}
}

// Name returns the name to use for running these tests.
func (config WithOptionsTestConfig[O, B, SO, SB]) Name() string {
	return "WithOptions"
}

// ExecuteTests runs the standard set of WithOptions tests for the configured resource.
func (config WithOptionsTestConfig[O, B, SO, SB]) ExecuteTests(t *testing.T) {
	t.Helper()

	t.Run("scheme attacher adds GVK", createSchemeAttacherGVKTest[O, SO](config.SchemeAttacher, config.ExpectedGVK))

	testCases := []struct {
		name         string
		builderError error
		options      []func(SB) (SB, error)
		assertError  func(error) bool
	}{
		{
			name:        "options are applied successfully",
			options:     []func(SB) (SB, error){testAnnotationOption[O, B, SO, SB]()},
			assertError: isErrorNil,
		},
		{
			name:         "invalid builder stops processing",
			builderError: errInvalidBuilder,
			options:      []func(SB) (SB, error){testAnnotationOption[O, B, SO, SB]()},
			assertError:  isInvalidBuilder,
		},
		{
			name:        "error in option is captured in builder",
			options:     []func(SB) (SB, error){testFailingOption[O, B, SO, SB]()},
			assertError: isOptionFailure,
		},
		{
			name:        "nil options are skipped",
			options:     []func(SB) (SB, error){nil, testAnnotationOption[O, B, SO, SB](), nil},
			assertError: isErrorNil,
		},
		{
			name:        "error stops subsequent options",
			options:     []func(SB) (SB, error){testFailingOption[O, B, SO, SB](), testAnnotationOption[O, B, SO, SB]()},
			assertError: isOptionFailure,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := clients.GetTestClients(clients.TestClientParams{
				SchemeAttachers: []clients.SchemeAttacher{config.SchemeAttacher},
			})

			var builder SB
			if config.ResourceScope.IsNamespaced() {
				builder = common.NewNamespacedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName, testResourceNamespace)
			} else {
				builder = common.NewClusterScopedBuilder[O, B, SO, SB](client, config.SchemeAttacher, testResourceName)
			}

			builder.SetError(testCase.builderError)

			result := config.withOptionsFunc(builder, testCase.options...)

			require.NotNil(t, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error, got: %v", result.GetError())

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
		})
	}
}
