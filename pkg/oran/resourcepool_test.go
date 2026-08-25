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
	testResourcePoolName      = "test-resourcepool"
	testResourcePoolNamespace = "test-namespace"
)

var resourcePoolGVK = inventoryv1alpha1.GroupVersion.WithKind("ResourcePool")

var (
	defaultResourcePoolCondition = metav1.Condition{
		Type:   inventoryv1alpha1.ConditionTypeReady,
		Status: metav1.ConditionTrue,
		Reason: inventoryv1alpha1.ReasonReady,
	}

	errResourcePoolNameEmpty = commonerrors.NewBuilderFieldEmpty(
		key.NewResourceKey("ResourcePool", "", testResourcePoolNamespace),
		commonerrors.BuilderFieldName,
	)
)

func TestNewResourcePoolBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewResourcePoolBuilder,
		inventoryv1alpha1.AddToScheme,
		resourcePoolGVK,
	).ExecuteTests(t)
}

func TestPullResourcePool(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullResourcePool,
		inventoryv1alpha1.AddToScheme,
		resourcePoolGVK,
	).ExecuteTests(t)
}

func TestListResourcePools(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListResourcePools,
		inventoryv1alpha1.AddToScheme,
		resourcePoolGVK,
	).ExecuteTests(t)
}

func TestListReadyResourcePools(t *testing.T) {
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
			name: "filters ready resourcepools",
			mockObjects: func() []runtime.Object {
				readyResourcePool := buildDummyResourcePool("ready-resourcepool", testResourcePoolNamespace)
				readyResourcePool.Status.Conditions = append(readyResourcePool.Status.Conditions, defaultResourcePoolCondition)

				return []runtime.Object{
					readyResourcePool,
					buildDummyResourcePool("not-ready-resourcepool", testResourcePoolNamespace),
				}
			}(),
			assertError: func(err error) bool {
				return err == nil
			},
			expectedCount: 1,
			expectedName:  "ready-resourcepool",
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

			resourcePools, err := ListReadyResourcePools(testSettings)
			require.True(t, testCase.assertError(err), "unexpected error: %v", err)
			require.Len(t, resourcePools, testCase.expectedCount)

			if testCase.expectedCount > 0 {
				assert.Equal(t, testCase.expectedName, resourcePools[0].Definition.Name)
				assert.Equal(t, testResourcePoolNamespace, resourcePools[0].Definition.Namespace)
			}
		})
	}
}

func TestResourcePoolMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[inventoryv1alpha1.ResourcePool, ResourcePoolBuilder](
		inventoryv1alpha1.AddToScheme,
		resourcePoolGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestResourcePoolWithOCloudSiteName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		builder        *ResourcePoolBuilder
		oCloudSiteName string
		expectedError  error
		expectedValue  string
	}{
		{
			name:           "sets oCloudSiteName on valid builder",
			builder:        newValidResourcePoolBuilder(newResourcePoolTestClient()),
			oCloudSiteName: testOCloudSiteName,
			expectedValue:  testOCloudSiteName,
		},
		{
			name:           "empty oCloudSiteName sets builder error",
			builder:        newValidResourcePoolBuilder(newResourcePoolTestClient()),
			oCloudSiteName: "",
			expectedError:  fmt.Errorf("resourcePool 'oCloudSiteName' cannot be empty"),
		},
		{
			name:           "invalid builder short circuits",
			builder:        newInvalidResourcePoolBuilder(newResourcePoolTestClient()),
			oCloudSiteName: testOCloudSiteName,
			expectedError:  errResourcePoolNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, testCase.builder)

			result := testCase.builder.WithOCloudSiteName(testCase.oCloudSiteName)
			require.Same(t, testCase.builder, result)
			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.expectedValue, result.Definition.Spec.OCloudSiteName)
			}
		})
	}
}

func TestResourcePoolWithDescription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       *ResourcePoolBuilder
		description   string
		expectedError error
		expectedValue string
	}{
		{
			name:          "sets description on valid builder",
			builder:       newValidResourcePoolBuilder(newResourcePoolTestClient()),
			description:   "test description",
			expectedValue: "test description",
		},
		{
			name:          "invalid builder short circuits",
			builder:       newInvalidResourcePoolBuilder(newResourcePoolTestClient()),
			description:   "test description",
			expectedError: errResourcePoolNameEmpty,
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

func TestResourcePoolWaitForCondition(t *testing.T) {
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
			name:         "resourcepool does not exist",
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
				resourcePool := buildDummyResourcePool(testResourcePoolName, testResourcePoolNamespace)
				if testCase.conditionMet {
					resourcePool.Status.Conditions = append(resourcePool.Status.Conditions, defaultResourcePoolCondition)
				}

				runtimeObjects = append(runtimeObjects, resourcePool)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: inventoryTestSchemes,
			})

			var testBuilder *ResourcePoolBuilder
			if testCase.valid {
				testBuilder = newValidResourcePoolBuilder(testSettings)
			} else {
				testBuilder = newInvalidResourcePoolBuilder(testSettings)
			}

			_, err := testBuilder.WaitForCondition(defaultResourcePoolCondition, time.Second)
			require.True(t, testCase.assertError(err), "unexpected error: %v", err)
		})
	}
}

// buildDummyResourcePool returns a ResourcePool with the provided name and namespace.
func buildDummyResourcePool(name, nsname string) *inventoryv1alpha1.ResourcePool {
	return &inventoryv1alpha1.ResourcePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// newResourcePoolTestClient returns a test client with the inventory scheme attached.
func newResourcePoolTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: inventoryTestSchemes,
	})
}

// newValidResourcePoolBuilder returns a valid ResourcePoolBuilder with default test name and namespace.
func newValidResourcePoolBuilder(apiClient *clients.Settings) *ResourcePoolBuilder {
	return NewResourcePoolBuilder(apiClient, testResourcePoolName, testResourcePoolNamespace)
}

// newInvalidResourcePoolBuilder returns a ResourcePoolBuilder with an empty name for validation testing.
func newInvalidResourcePoolBuilder(apiClient *clients.Settings) *ResourcePoolBuilder {
	return NewResourcePoolBuilder(apiClient, "", testResourcePoolNamespace)
}
