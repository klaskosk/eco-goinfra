package clusterlogging

import (
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	lokiv1 "github.com/grafana/loki/operator/apis/loki/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultLokiStackName      = "lokistack-test"
	defaultLokiStackNamespace = "lokistack-space"
	lokiStackComponents       = &lokiv1.LokiComponentSpec{
		NodeSelector: map[string]string{
			"node-role.kubernetes.io/infra": "",
		},
		Tolerations: []corev1.Toleration{{
			Key:      "node-role.kubernetes.io/infra",
			Operator: "Exists",
		}},
	}
	lokiv1TestSchemes = []clients.SchemeAttacher{
		lokiv1.AddToScheme,
	}
)

func TestPullLokiStack(t *testing.T) {
	generateLokiStack := func(name, namespace string) *lokiv1.LokiStack {
		return &lokiv1.LokiStack{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                defaultLokiStackName,
			namespace:           defaultLokiStackNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultLokiStackNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("lokiStack 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultLokiStackName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("lokiStack 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "lokitest",
			namespace:           defaultLokiStackNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("lokiStack object lokitest does not exist " +
				"in namespace lokistack-space"),
			client: true,
		},
		{
			name:                "triggerauthtest",
			namespace:           defaultLokiStackNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("lokiStack 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testLokiStack := generateLokiStack(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testLokiStack)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: lokiv1TestSchemes,
			})
		}

		builderResult, err := PullLokiStack(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testLokiStack.Name, builderResult.Object.Name)
			assert.Equal(t, testLokiStack.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewLokiStackBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultLokiStackName,
			namespace:     defaultLokiStackNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultLokiStackNamespace,
			expectedError: "lokiStack 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultLokiStackName,
			namespace:     "",
			expectedError: "lokiStack 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultLokiStackName,
			namespace:     defaultLokiStackNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testLokiStackBuilder := NewLokiStackBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testLokiStackBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testLokiStackBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testLokiStackBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testLokiStackBuilder.errorMsg)
			assert.NotNil(t, testLokiStackBuilder.Definition)
		}
	}
}

func TestLokiStackExists(t *testing.T) {
	testCases := []struct {
		testLokiStack  *LokiStackBuilder
		expectedStatus bool
	}{
		{
			testLokiStack:  buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testLokiStack:  buildInValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testLokiStack:  buildValidLokiStackBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testLokiStack.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestLokiStackGet(t *testing.T) {
	testCases := []struct {
		testLokiStack *LokiStackBuilder
		expectedError error
	}{
		{
			testLokiStack: buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testLokiStack: buildInValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: fmt.Errorf("lokiStack 'name' cannot be empty"),
		},
		{
			testLokiStack: buildValidLokiStackBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("lokistacks.loki.grafana.com \"lokistack-test\" not found"),
		},
	}

	for _, testCase := range testCases {
		lokiStackObj, err := testCase.testLokiStack.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, lokiStackObj.Name, testCase.testLokiStack.Definition.Name)
			assert.Equal(t, lokiStackObj.Namespace, testCase.testLokiStack.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestLokiStackCreate(t *testing.T) {
	testCases := []struct {
		testLokiStack *LokiStackBuilder
		expectedError string
	}{
		{
			testLokiStack: buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: "",
		},
		{
			testLokiStack: buildInValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: "lokiStack 'name' cannot be empty",
		},
		{
			testLokiStack: buildValidLokiStackBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testLokiStackBuilder, err := testCase.testLokiStack.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testLokiStackBuilder.Definition.Name, testLokiStackBuilder.Object.Name)
			assert.Equal(t, testLokiStackBuilder.Definition.Namespace, testLokiStackBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestLokiStackDelete(t *testing.T) {
	testCases := []struct {
		testLokiStack *LokiStackBuilder
		expectedError error
	}{
		{
			testLokiStack: buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testLokiStack: buildValidLokiStackBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testLokiStack.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testLokiStack.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestLokiStackUpdate(t *testing.T) {
	testCases := []struct {
		testLokiStack *LokiStackBuilder
		testSize      lokiv1.LokiStackSizeType
		expectedError string
	}{
		{
			testLokiStack: buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			testSize:      lokiv1.SizeOneXDemo,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, lokiv1.LokiStackSizeType(""), testCase.testLokiStack.Definition.Spec.Size)
		assert.Nil(t, nil, testCase.testLokiStack.Object)
		testCase.testLokiStack.WithSize(testCase.testSize)
		_, err := testCase.testLokiStack.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testSize, testCase.testLokiStack.Definition.Spec.Size)
		}
	}
}

func TestLokiStackWithSize(t *testing.T) {
	testCases := []struct {
		testSize      lokiv1.LokiStackSizeType
		expectedError string
	}{
		{
			testSize:      lokiv1.SizeOneXDemo,
			expectedError: "",
		},
		{
			testSize:      lokiv1.SizeOneXSmall,
			expectedError: "",
		},
		{
			testSize:      lokiv1.SizeOneXMedium,
			expectedError: "",
		},
		{
			testSize:      lokiv1.SizeOneXExtraSmall,
			expectedError: "",
		},
		{
			testSize:      "",
			expectedError: "'size' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithSize(testCase.testSize)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testSize, result.Definition.Spec.Size)
		}
	}
}

func TestLokiStackWithStorage(t *testing.T) {
	testCases := []struct {
		testStorage   lokiv1.ObjectStorageSpec
		expectedError string
	}{
		{
			testStorage: lokiv1.ObjectStorageSpec{
				Secret: lokiv1.ObjectStorageSecretSpec{
					Type: "s3",
					Name: "test",
				},
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithStorage(testCase.testStorage)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testStorage, result.Definition.Spec.Storage)
		}
	}
}

func TestLokiStackWithStorageClassName(t *testing.T) {
	testCases := []struct {
		testStorageClassName string
		expectedError        string
	}{
		{
			testStorageClassName: "gp2",
			expectedError:        "",
		},
		{
			testStorageClassName: "",
			expectedError:        "'storageClassName' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithStorageClassName(testCase.testStorageClassName)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testStorageClassName, result.Definition.Spec.StorageClassName)
		}
	}
}

func TestLokiStackWithTenants(t *testing.T) {
	testCases := []struct {
		testTenants   lokiv1.TenantsSpec
		expectedError string
	}{
		{
			testTenants: lokiv1.TenantsSpec{
				Mode: "openshift-logging",
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithTenants(testCase.testTenants)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testTenants.Mode, result.Definition.Spec.Tenants.Mode)
		}
	}
}

func TestLokiStackWithRules(t *testing.T) {
	testCases := []struct {
		testRules     lokiv1.RulesSpec
		expectedError string
	}{
		{
			testRules: lokiv1.RulesSpec{
				Enabled: true,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"openshift.io/cluster-monitoring": "true"},
				},
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"openshift.io/cluster-monitoring": "true"},
				},
			},
			expectedError: "",
		},
		{
			testRules: lokiv1.RulesSpec{
				Enabled: false,
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithRules(testCase.testRules)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testRules.Enabled, result.Definition.Spec.Rules.Enabled)

			if testCase.testRules.Enabled {
				assert.Equal(t, testCase.testRules.Selector, result.Definition.Spec.Rules.Selector)
				assert.Equal(t, testCase.testRules.NamespaceSelector, result.Definition.Spec.Rules.NamespaceSelector)
			}
		}
	}
}

func TestLokiStackWithManagementState(t *testing.T) {
	testCases := []struct {
		testManagementState lokiv1.ManagementStateType
		expectedError       string
	}{
		{
			testManagementState: lokiv1.ManagementStateManaged,
			expectedError:       "",
		},
		{
			testManagementState: lokiv1.ManagementStateUnmanaged,
			expectedError:       "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithManagementState(testCase.testManagementState)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testManagementState, result.Definition.Spec.ManagementState)
		}
	}
}

func TestLokiStackWithLimits(t *testing.T) {
	testCases := []struct {
		testLimitSpec lokiv1.LimitsSpec
		expectedError string
	}{
		{
			testLimitSpec: lokiv1.LimitsSpec{
				Global: &lokiv1.LimitsTemplateSpec{
					Retention: &lokiv1.RetentionLimitSpec{
						Days: 7,
					},
				},
			},
			expectedError: "",
		},
		{
			testLimitSpec: lokiv1.LimitsSpec{},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithLimits(testCase.testLimitSpec)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testLimitSpec.Global, result.Definition.Spec.Limits.Global)
		}
	}
}

func TestLokiStackWithTemplate(t *testing.T) {
	testCases := []struct {
		testTemplate  lokiv1.LokiTemplateSpec
		expectedError string
	}{
		{
			testTemplate: lokiv1.LokiTemplateSpec{
				Compactor:     lokiStackComponents,
				Distributor:   lokiStackComponents,
				Ingester:      lokiStackComponents,
				Querier:       lokiStackComponents,
				QueryFrontend: lokiStackComponents,
				Gateway:       lokiStackComponents,
				IndexGateway:  lokiStackComponents,
				Ruler:         lokiStackComponents,
			},
			expectedError: "",
		},
		{
			testTemplate:  lokiv1.LokiTemplateSpec{},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithTemplate(testCase.testTemplate)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testTemplate.Compactor, result.Definition.Spec.Template.Compactor)
			assert.Equal(t, testCase.testTemplate.Distributor, result.Definition.Spec.Template.Distributor)
			assert.Equal(t, testCase.testTemplate.Ingester, result.Definition.Spec.Template.Ingester)
			assert.Equal(t, testCase.testTemplate.Querier, result.Definition.Spec.Template.Querier)
			assert.Equal(t, testCase.testTemplate.QueryFrontend, result.Definition.Spec.Template.QueryFrontend)
			assert.Equal(t, testCase.testTemplate.Gateway, result.Definition.Spec.Template.Gateway)
			assert.Equal(t, testCase.testTemplate.IndexGateway, result.Definition.Spec.Template.IndexGateway)
			assert.Equal(t, testCase.testTemplate.Ruler, result.Definition.Spec.Template.Ruler)
		}
	}
}

func TestLokiStackIsReady(t *testing.T) {
	testCases := []struct {
		testLokiStack *LokiStackBuilder
		testCondition bool
	}{
		{
			testLokiStack: buildValidLokiStackBuilderWithCondition(buildLokiStackClientWithDummyObject(), "Ready"),
			testCondition: true,
		},
		{
			testLokiStack: buildValidLokiStackBuilderWithCondition(buildLokiStackClientWithDummyObject(), "NotReady"),
			testCondition: false,
		},
		{
			testLokiStack: buildValidLokiStackBuilderWithCondition(clients.GetTestClients(clients.TestClientParams{}),
				"Ready"),
			testCondition: false,
		},
	}

	for _, testCase := range testCases {
		isReadyResult := testCase.testLokiStack.IsReady(2 * time.Second)

		assert.Equal(t, testCase.testCondition, isReadyResult)
	}
}

func buildValidLokiStackBuilderWithCondition(apiClient *clients.Settings,
	conditionType string) *LokiStackBuilder {
	lokiStackBuilder := buildValidLokiStackBuilder(apiClient)
	lokiStackBuilder.Definition.Status.Conditions = []metav1.Condition{{
		Type:   conditionType,
		Status: metav1.ConditionTrue,
	}}

	return lokiStackBuilder
}

func buildValidLokiStackBuilder(apiClient *clients.Settings) *LokiStackBuilder {
	lokiStackBuilder := NewLokiStackBuilder(
		apiClient, defaultLokiStackName, defaultLokiStackNamespace)

	return lokiStackBuilder
}

func buildInValidLokiStackBuilder(apiClient *clients.Settings) *LokiStackBuilder {
	lokiStackBuilder := NewLokiStackBuilder(
		apiClient, "", defaultLokiStackNamespace)

	return lokiStackBuilder
}

func buildLokiStackClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyTriggerAuthentication(),
		SchemeAttachers: lokiv1TestSchemes,
	})
}

func buildDummyTriggerAuthentication() []runtime.Object {
	return append([]runtime.Object{}, &lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultLokiStackName,
			Namespace: defaultLokiStackNamespace,
		},
	})
}
