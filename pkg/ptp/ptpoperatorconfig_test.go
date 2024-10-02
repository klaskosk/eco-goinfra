package ptp

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	ptpv1 "github.com/openshift/ptp-operator/api/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultPOCName      = "default"
	defaultPOCNamespace = "openshift-ptp"
)

func TestPullPtpOperatorConfig(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultPOCName,
			nsname:              defaultPOCNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultPOCNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("ptpOperatorConfig 'name' cannot be empty"),
		},
		{
			name:                defaultPOCName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("ptpOperatorConfig 'nsname' cannot be empty"),
		},
		{
			name:                defaultPOCName,
			nsname:              defaultPOCNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"ptpOperatorConfig object %s does not exist in namespace %s", defaultPOCName, defaultPOCNamespace),
		},
		{
			name:                defaultPOCName,
			nsname:              defaultPOCNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("ptpOperatorConfig 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPOC := buildDummpPOC(testCase.name, testCase.nsname)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPOC)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullPtpOperatorConfig(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testPOC.Name, testBuilder.Definition.Name)
			assert.Equal(t, testPOC.Namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestPtpOperatorConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *PtpOperatorConfigBuilder
		expectedError string
	}{
		{
			testBuilder:   newPtpOperatorConfig(buildTestClientWithDummyPOC()),
			expectedError: "",
		},
		{
			testBuilder:   newPtpOperatorConfig(buildTestClientWithPtpScheme()),
			expectedError: "ptpoperatorconfigs.ptp.openshift.io \"default\" not found",
		},
	}

	for _, testCase := range testCases {
		ptpOperatorConfig, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, ptpOperatorConfig.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, ptpOperatorConfig.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestPtpOperatorConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PtpOperatorConfigBuilder
		exists      bool
	}{
		{
			testBuilder: newPtpOperatorConfig(buildTestClientWithDummyPOC()),
			exists:      true,
		},
		{
			testBuilder: newPtpOperatorConfig(buildTestClientWithPtpScheme()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPtpOperatorConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *PtpOperatorConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   newPtpOperatorConfig(buildTestClientWithPtpScheme()),
			expectedError: nil,
		},
		{
			testBuilder:   newPtpOperatorConfig(buildTestClientWithDummyPOC()),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
			assert.Equal(t, testBuilder.Definition.Namespace, testBuilder.Object.Namespace)
		}
	}
}

func TestPtpOperatorConfigUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists bool
		expectedError error
	}{
		{
			alreadyExists: false,
			expectedError: fmt.Errorf("cannot update non-existent ptpOperatorConfig"),
		},
		{
			alreadyExists: true,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithPtpScheme()

		if testCase.alreadyExists {
			testSettings = buildTestClientWithDummyPOC()
		}

		testBuilder := newPtpOperatorConfig(testSettings)
		assert.Nil(t, testBuilder.Definition.Spec.EventConfig)

		testBuilder.Definition.Spec.EventConfig = &ptpv1.PtpEventConfig{}

		testBuilder, err := testBuilder.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, testBuilder.Object.Spec.EventConfig)
		}
	}
}

func TestPtpOperatorConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PtpOperatorConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   newPtpOperatorConfig(buildTestClientWithDummyPOC()),
			expectedError: nil,
		},
		{
			testBuilder:   newPtpOperatorConfig(buildTestClientWithPtpScheme()),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestPtpOperatorConfigValidate(t *testing.T) {
	testCases := []struct {
		builderNil      bool
		definitionNil   bool
		apiClientNil    bool
		builderErrorMsg string
		expectedError   error
	}{
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   nil,
		},
		{
			builderNil:      true,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("error: received nil ptpOperatorConfig builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined ptpOperatorConfig"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("ptpOperatorConfig builder cannot have nil apiClient"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "test error",
			expectedError:   fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		testBuilder := newPtpOperatorConfig(buildTestClientWithPtpScheme())

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			testBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := testBuilder.validate()
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, valid)
	}
}

// buildDummpPOC returns a PtpOperatorConfig with the provided name and namespace.
func buildDummpPOC(name, namespace string) *ptpv1.PtpOperatorConfig {
	return &ptpv1.PtpOperatorConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// buildTestClientWithDummyPOC returns a client with a mock PtpOperatorConfig.
func buildTestClientWithDummyPOC() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummpPOC(defaultPOCName, defaultPOCNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

func newPtpOperatorConfig(apiClient *clients.Settings) *PtpOperatorConfigBuilder {
	return &PtpOperatorConfigBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummpPOC(defaultPOCName, defaultPOCNamespace),
	}
}
