package route

import (
	"fmt"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
)

func Test_Builder_Pull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull,
		routev1.AddToScheme,
		routev1.GroupVersion.WithKind("Route"),
	).ExecuteTests(t)
}

func Test_Builder_Methods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[routev1.Route, Builder](
		routev1.AddToScheme,
		routev1.GroupVersion.WithKind("Route"),
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		Run(t)
}

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	t.Run("common namespaced builder behavior", func(t *testing.T) {
		t.Parallel()

		testhelper.NewNamespacedBuilderTestConfig(
			func(apiClient *clients.Settings, name, nsname string) *Builder {
				return NewBuilder(apiClient, name, nsname, "route-test-service")
			},
			routev1.AddToScheme,
			routev1.GroupVersion.WithKind("Route"),
		).ExecuteTests(t)
	})

	testCases := []struct {
		name          string
		targetService string
		expectedError error
	}{
		{
			name:          "valid service name sets target reference",
			targetService: "route-test-service",
			expectedError: nil,
		},
		{
			name:          "empty service name returns error",
			targetService: "",
			expectedError: fmt.Errorf("route 'serviceName' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := NewBuilder(
				clients.GetTestClients(clients.TestClientParams{}),
				"route-test-name",
				"route-test-namespace",
				testCase.targetService,
			)

			err := testBuilder.GetError()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.targetService, testBuilder.Definition.Spec.To.Name)
				assert.Equal(t, "Service", testBuilder.Definition.Spec.To.Kind)
			}
		})
	}
}

func TestWithTargetPortNumber(t *testing.T) {
	t.Parallel()

	t.Run("valid builder sets target port by number", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildValidRouteTestBuilder(getTestRouteAPIClient())
		testBuilder.WithTargetPortNumber(8080)

		assert.NoError(t, testBuilder.GetError())
		assert.Equal(t, int32(8080), testBuilder.Definition.Spec.Port.TargetPort.IntVal)
	})

	t.Run("builder with error short-circuits without setting port", func(t *testing.T) {
		t.Parallel()

		testBuilder := buildInvalidRouteTestBuilder(getTestRouteAPIClient())
		testBuilder.WithTargetPortNumber(8080)

		assert.Equal(t, fmt.Errorf("route 'serviceName' cannot be empty"), testBuilder.GetError())
		assert.Nil(t, testBuilder.Definition.Spec.Port)
	})
}

func TestWithTargetPortName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		portName      string
		expectedError error
	}{
		{
			name:          "empty port name returns error",
			portName:      "",
			expectedError: fmt.Errorf("route target port name cannot be empty string"),
		},
		{
			name:          "non-empty port name sets target reference",
			portName:      "8080-target",
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidRouteTestBuilder(getTestRouteAPIClient())
			testBuilder.WithTargetPortName(testCase.portName)

			err := testBuilder.GetError()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.portName, testBuilder.Definition.Spec.Port.TargetPort.StrVal)
			}
		})
	}
}

func TestWithHostDomain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		hostDomain    string
		expectedError error
	}{
		{
			name:          "empty host domain returns error",
			hostDomain:    "",
			expectedError: fmt.Errorf("route host domain cannot be empty string"),
		},
		{
			name:          "non-empty host domain sets spec host",
			hostDomain:    "app.demo-server.dummy.domain.com",
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidRouteTestBuilder(getTestRouteAPIClient())
			testBuilder.WithHostDomain(testCase.hostDomain)

			err := testBuilder.GetError()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.hostDomain, testBuilder.Definition.Spec.Host)
			}
		})
	}
}

func TestWithWildCardPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		policy        string
		expectedError error
	}{
		{
			name:          "empty policy is unsupported",
			policy:        "",
			expectedError: getUnsupportedWildCardPoliciesError(),
		},
		{
			name:          "Any policy is unsupported",
			policy:        "Any",
			expectedError: getUnsupportedWildCardPoliciesError(),
		},
		{
			name:          "Subdomain policy is supported",
			policy:        "Subdomain",
			expectedError: nil,
		},
		{
			name:          "None policy is supported",
			policy:        "None",
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testBuilder := buildValidRouteTestBuilder(getTestRouteAPIClient())
			testBuilder.WithWildCardPolicy(testCase.policy)

			err := testBuilder.GetError()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.Equal(t, routev1.WildcardPolicyType(testCase.policy), testBuilder.Definition.Spec.WildcardPolicy)
			}
		})
	}
}

// getTestRouteAPIClient returns a test client for the Route resource. It has no mock objects.
func getTestRouteAPIClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}

// buildValidRouteTestBuilder returns a valid Builder for testing purposes.
func buildValidRouteTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "route-test-name", "route-test-namespace", "route-test-service")
}

// buildInvalidRouteTestBuilder returns a Builder that is already in an error state (empty service name).
func buildInvalidRouteTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "route-test-name", "route-test-namespace", "")
}
