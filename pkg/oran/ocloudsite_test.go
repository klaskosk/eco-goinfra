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
	testOCloudSiteName      = "test-ocloudsite"
	testOCloudSiteNamespace = "test-namespace"
	testGlobalLocationName  = "test-location"
)

var ocloudSiteGVK = inventoryv1alpha1.GroupVersion.WithKind("OCloudSite")

var (
	defaultOCloudSiteCondition = metav1.Condition{
		Type:   inventoryv1alpha1.ConditionTypeReady,
		Status: metav1.ConditionTrue,
		Reason: inventoryv1alpha1.ReasonReady,
	}

	errOCloudSiteNameEmpty = commonerrors.NewBuilderFieldEmpty(
		key.NewResourceKey("OCloudSite", "", testOCloudSiteNamespace),
		commonerrors.BuilderFieldName,
	)
)

func TestNewOCloudSiteBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewOCloudSiteBuilder,
		inventoryv1alpha1.AddToScheme,
		ocloudSiteGVK,
	).ExecuteTests(t)
}

func TestPullOCloudSite(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullOCloudSite,
		inventoryv1alpha1.AddToScheme,
		ocloudSiteGVK,
	).ExecuteTests(t)
}

func TestListOCloudSites(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListOCloudSites,
		inventoryv1alpha1.AddToScheme,
		ocloudSiteGVK,
	).ExecuteTests(t)
}

func TestListReadyOCloudSites(t *testing.T) {
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
			name: "filters ready ocloudsites",
			mockObjects: func() []runtime.Object {
				readyOCloudSite := buildDummyOCloudSite("ready-ocloudsite", testOCloudSiteNamespace)
				readyOCloudSite.Status.Conditions = append(readyOCloudSite.Status.Conditions, defaultOCloudSiteCondition)

				return []runtime.Object{
					readyOCloudSite,
					buildDummyOCloudSite("not-ready-ocloudsite", testOCloudSiteNamespace),
				}
			}(),
			assertError: func(err error) bool {
				return err == nil
			},
			expectedCount: 1,
			expectedName:  "ready-ocloudsite",
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

			ocloudSites, err := ListReadyOCloudSites(testSettings)
			require.True(t, testCase.assertError(err), "unexpected error: %v", err)
			require.Len(t, ocloudSites, testCase.expectedCount)

			if testCase.expectedCount > 0 {
				assert.Equal(t, testCase.expectedName, ocloudSites[0].Definition.Name)
				assert.Equal(t, testOCloudSiteNamespace, ocloudSites[0].Definition.Namespace)
			}
		})
	}
}

func TestOCloudSiteMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[inventoryv1alpha1.OCloudSite, OCloudSiteBuilder](
		inventoryv1alpha1.AddToScheme,
		ocloudSiteGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

func TestOCloudSiteWithGlobalLocationName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		builder            *OCloudSiteBuilder
		globalLocationName string
		expectedError      error
		expectedValue      string
	}{
		{
			name:               "sets globalLocationName on valid builder",
			builder:            newValidOCloudSiteBuilder(newOCloudSiteTestClient()),
			globalLocationName: testGlobalLocationName,
			expectedValue:      testGlobalLocationName,
		},
		{
			name:               "empty globalLocationName sets builder error",
			builder:            newValidOCloudSiteBuilder(newOCloudSiteTestClient()),
			globalLocationName: "",
			expectedError:      fmt.Errorf("oCloudSite 'globalLocationName' cannot be empty"),
		},
		{
			name:               "invalid builder short circuits",
			builder:            newInvalidOCloudSiteBuilder(newOCloudSiteTestClient()),
			globalLocationName: testGlobalLocationName,
			expectedError:      errOCloudSiteNameEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, testCase.builder)

			result := testCase.builder.WithGlobalLocationName(testCase.globalLocationName)
			require.Same(t, testCase.builder, result)
			assert.Equal(t, testCase.expectedError, result.GetError())

			if testCase.expectedError == nil {
				assert.Equal(t, testCase.expectedValue, result.Definition.Spec.GlobalLocationName)
			}
		})
	}
}

func TestOCloudSiteWithDescription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       *OCloudSiteBuilder
		description   string
		expectedError error
		expectedValue string
	}{
		{
			name:          "sets description on valid builder",
			builder:       newValidOCloudSiteBuilder(newOCloudSiteTestClient()),
			description:   "test description",
			expectedValue: "test description",
		},
		{
			name:          "invalid builder short circuits",
			builder:       newInvalidOCloudSiteBuilder(newOCloudSiteTestClient()),
			description:   "test description",
			expectedError: errOCloudSiteNameEmpty,
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

func TestOCloudSiteWaitForCondition(t *testing.T) {
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
			name:         "ocloudsite does not exist",
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
				ocloudSite := buildDummyOCloudSite(testOCloudSiteName, testOCloudSiteNamespace)
				if testCase.conditionMet {
					ocloudSite.Status.Conditions = append(ocloudSite.Status.Conditions, defaultOCloudSiteCondition)
				}

				runtimeObjects = append(runtimeObjects, ocloudSite)
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: inventoryTestSchemes,
			})

			var testBuilder *OCloudSiteBuilder
			if testCase.valid {
				testBuilder = newValidOCloudSiteBuilder(testSettings)
			} else {
				testBuilder = newInvalidOCloudSiteBuilder(testSettings)
			}

			_, err := testBuilder.WaitForCondition(defaultOCloudSiteCondition, time.Second)
			require.True(t, testCase.assertError(err), "unexpected error: %v", err)
		})
	}
}

// buildDummyOCloudSite returns an OCloudSite with the provided name and namespace.
func buildDummyOCloudSite(name, nsname string) *inventoryv1alpha1.OCloudSite {
	return &inventoryv1alpha1.OCloudSite{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// newOCloudSiteTestClient returns a test client with the inventory scheme attached.
func newOCloudSiteTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: inventoryTestSchemes,
	})
}

// newValidOCloudSiteBuilder returns a valid OCloudSiteBuilder with default test name and namespace.
func newValidOCloudSiteBuilder(apiClient *clients.Settings) *OCloudSiteBuilder {
	return NewOCloudSiteBuilder(apiClient, testOCloudSiteName, testOCloudSiteNamespace)
}

// newInvalidOCloudSiteBuilder returns an OCloudSiteBuilder with an empty name for validation testing.
func newInvalidOCloudSiteBuilder(apiClient *clients.Settings) *OCloudSiteBuilder {
	return NewOCloudSiteBuilder(apiClient, "", testOCloudSiteNamespace)
}
