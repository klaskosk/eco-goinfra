package oran

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	inventoryv1alpha1 "github.com/openshift-kni/oran-o2ims/api/inventory/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	testLocationName      = "test-location"
	testLocationNamespace = "test-namespace"
)

var locationGVK = inventoryv1alpha1.GroupVersion.WithKind("Location")

var (
	defaultLocationCondition = metav1.Condition{
		Type:   inventoryv1alpha1.ConditionTypeReady,
		Status: metav1.ConditionTrue,
		Reason: inventoryv1alpha1.ReasonReady,
	}

	inventoryTestSchemes = []clients.SchemeAttacher{
		inventoryv1alpha1.AddToScheme,
	}

	errLocationNameEmpty = commonerrors.NewBuilderFieldEmpty(
		key.NewResourceKey("Location", "", testLocationNamespace),
		commonerrors.BuilderFieldName,
	)
)

func TestNewLocationBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewLocationBuilder,
		inventoryv1alpha1.AddToScheme,
		locationGVK,
	).ExecuteTests(t)
}

func TestPullLocation(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullLocation,
		inventoryv1alpha1.AddToScheme,
		locationGVK,
	).ExecuteTests(t)
}

func TestListLocations(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListLocations,
		inventoryv1alpha1.AddToScheme,
		locationGVK,
	).ExecuteTests(t)
}

func TestListReadyLocations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		mockObjects      []runtime.Object
		interceptorFuncs interceptor.Funcs
		assertError      func(error) bool
		expectedCount    int
		expectedName     string
	}{
		{
			name: "filters ready locations",
			mockObjects: func() []runtime.Object {
				readyLocation := buildDummyLocation("ready-location", testLocationNamespace)
				readyLocation.Status.Conditions = append(readyLocation.Status.Conditions, defaultLocationCondition)

				return []runtime.Object{
					readyLocation,
					buildDummyLocation("not-ready-location", testLocationNamespace),
				}
			}(),
			assertError: func(err error) bool {
				return err == nil
			},
			expectedCount: 1,
			expectedName:  "ready-location",
		},
		{
			name: "list failure returns error",
			interceptorFuncs: interceptor.Funcs{
				List: func(
					_ context.Context,
					_ runtimeclient.WithWatch,
					_ runtimeclient.ObjectList,
					_ ...runtimeclient.ListOption,
				) error {
					return errors.New("simulated list failure")
				},
			},
			assertError: func(err error) bool {
				return commonerrors.IsAPICallFailedWithVerb(err, "list")
			},
			expectedCount: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:   testCase.mockObjects,
				SchemeAttachers:  inventoryTestSchemes,
				InterceptorFuncs: testCase.interceptorFuncs,
			})

			locations, err := ListReadyLocations(testSettings)
			require.True(t, testCase.assertError(err), "unexpected error: %v", err)
			require.Len(t, locations, testCase.expectedCount)

			if testCase.expectedCount > 0 {
				assert.Equal(t, testCase.expectedName, locations[0].Definition.Name)
				assert.Equal(t, testLocationNamespace, locations[0].Definition.Namespace)
			}
		})
	}
}

func TestLocationMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[inventoryv1alpha1.Location, LocationBuilder](
		inventoryv1alpha1.AddToScheme,
		locationGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestLocationWithDescription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       *LocationBuilder
		description   string
		expectedError error
		expectedValue string
	}{
		{
			name:          "sets description on valid builder",
			builder:       newValidLocationBuilder(newLocationTestClient()),
			description:   "test description",
			expectedValue: "test description",
		},
		{
			name:          "invalid builder short circuits",
			builder:       newInvalidLocationBuilder(newLocationTestClient()),
			description:   "test description",
			expectedError: errLocationNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, testCase.builder)

			result := testCase.builder.WithDescription(testCase.description)
			require.Same(t, testCase.builder, result)
			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.expectedValue, result.Definition.Spec.Description)
			}
		})
	}
}

func TestLocationWithAddress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       *LocationBuilder
		address       string
		expectedError error
	}{
		{
			name:    "sets address on valid builder",
			builder: newValidLocationBuilder(newLocationTestClient()),
			address: "123 Main St",
		},
		{
			name:          "empty address sets builder error",
			builder:       newValidLocationBuilder(newLocationTestClient()),
			address:       "",
			expectedError: fmt.Errorf("location 'address' cannot be empty"),
		},
		{
			name:          "invalid builder short circuits",
			builder:       newInvalidLocationBuilder(newLocationTestClient()),
			address:       "123 Main St",
			expectedError: errLocationNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, testCase.builder)

			result := testCase.builder.WithAddress(testCase.address)
			require.Same(t, testCase.builder, result)
			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				require.NotNil(t, result.Definition.Spec.Address)
				assert.Equal(t, testCase.address, *result.Definition.Spec.Address)
			}
		})
	}
}

func TestLocationWithCoordinate(t *testing.T) {
	t.Parallel()

	validCoordinate := inventoryv1alpha1.GeoLocation{
		Latitude:  "40.7128",
		Longitude: "-74.0060",
	}

	testCases := []struct {
		name          string
		builder       *LocationBuilder
		coordinate    inventoryv1alpha1.GeoLocation
		expectedError error
	}{
		{
			name:       "sets coordinate on valid builder",
			builder:    newValidLocationBuilder(newLocationTestClient()),
			coordinate: validCoordinate,
		},
		{
			name:          "empty latitude sets builder error",
			builder:       newValidLocationBuilder(newLocationTestClient()),
			coordinate:    inventoryv1alpha1.GeoLocation{Longitude: "-74.0060"},
			expectedError: fmt.Errorf("location coordinate 'latitude' cannot be empty"),
		},
		{
			name:          "empty longitude sets builder error",
			builder:       newValidLocationBuilder(newLocationTestClient()),
			coordinate:    inventoryv1alpha1.GeoLocation{Latitude: "40.7128"},
			expectedError: fmt.Errorf("location coordinate 'longitude' cannot be empty"),
		},
		{
			name:          "invalid builder short circuits",
			builder:       newInvalidLocationBuilder(newLocationTestClient()),
			coordinate:    validCoordinate,
			expectedError: errLocationNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, testCase.builder)

			result := testCase.builder.WithCoordinate(testCase.coordinate)
			require.Same(t, testCase.builder, result)
			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				require.NotNil(t, result.Definition.Spec.Coordinate)
				assert.Equal(t, testCase.coordinate, *result.Definition.Spec.Coordinate)
			}
		})
	}
}

func TestLocationWithCivicAddress(t *testing.T) {
	t.Parallel()

	validElements := []inventoryv1alpha1.CivicAddressElement{
		{CaType: 0, CaValue: "US"},
		{CaType: 3, CaValue: "New York"},
	}

	testCases := []struct {
		name          string
		builder       *LocationBuilder
		elements      []inventoryv1alpha1.CivicAddressElement
		expectedError error
	}{
		{
			name:     "sets civic address on valid builder",
			builder:  newValidLocationBuilder(newLocationTestClient()),
			elements: validElements,
		},
		{
			name:          "empty civic address sets builder error",
			builder:       newValidLocationBuilder(newLocationTestClient()),
			elements:      []inventoryv1alpha1.CivicAddressElement{},
			expectedError: fmt.Errorf("location 'civicAddress' cannot be empty"),
		},
		{
			name:          "invalid builder short circuits",
			builder:       newInvalidLocationBuilder(newLocationTestClient()),
			elements:      validElements,
			expectedError: errLocationNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, testCase.builder)

			result := testCase.builder.WithCivicAddress(testCase.elements...)
			require.Same(t, testCase.builder, result)
			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.elements, result.Definition.Spec.CivicAddress)
			}
		})
	}
}

func TestLocationWaitForCondition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		conditionMet bool
		exists       bool
		valid        bool
		assertError  func(error) bool
	}{
		{
			name:         "condition met",
			conditionMet: true,
			exists:       true,
			valid:        true,
			assertError:  func(err error) bool { return err == nil },
		},
		{
			name:         "condition not met",
			conditionMet: false,
			exists:       true,
			valid:        true,
			assertError:  func(err error) bool { return errors.Is(err, context.DeadlineExceeded) },
		},
		{
			name:         "location does not exist",
			conditionMet: true,
			exists:       false,
			valid:        true,
			assertError: func(err error) bool {
				return err != nil && k8serrors.IsNotFound(err)
			},
		},
		{
			name:        "invalid builder",
			exists:      true,
			valid:       false,
			assertError: commonerrors.IsBuilderNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.exists {
				location := buildDummyLocation(testLocationName, testLocationNamespace)
				if testCase.conditionMet {
					location.Status.Conditions = append(location.Status.Conditions, defaultLocationCondition)
				}

				runtimeObjects = append(runtimeObjects, location)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: inventoryTestSchemes,
			})

			var testBuilder *LocationBuilder
			if testCase.valid {
				testBuilder = newValidLocationBuilder(testSettings)
			} else {
				testBuilder = newInvalidLocationBuilder(testSettings)
			}

			_, err := testBuilder.WaitForCondition(defaultLocationCondition, time.Second)
			require.True(t, testCase.assertError(err), "unexpected error: %v", err)
		})
	}
}

// buildDummyLocation returns a Location with the provided name and namespace.
func buildDummyLocation(name, nsname string) *inventoryv1alpha1.Location {
	return &inventoryv1alpha1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// newLocationTestClient returns a test client with the inventory scheme attached.
func newLocationTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: inventoryTestSchemes,
	})
}

// newValidLocationBuilder returns a valid LocationBuilder with default test name and namespace.
func newValidLocationBuilder(apiClient *clients.Settings) *LocationBuilder {
	return NewLocationBuilder(apiClient, testLocationName, testLocationNamespace)
}

// newInvalidLocationBuilder returns a LocationBuilder with an empty name for validation testing.
func newInvalidLocationBuilder(apiClient *clients.Settings) *LocationBuilder {
	return NewLocationBuilder(apiClient, "", testLocationNamespace)
}
