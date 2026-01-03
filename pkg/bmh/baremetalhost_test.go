package bmh

import (
	"fmt"
	"testing"
	"time"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultBmHostName       = "metallbio"
	defaultBmHostNsName     = "test-namespace"
	defaultBmHostAddress    = "1.1.1.1"
	defaultBmHostSecretName = "testsecret"
	defaultBmHostMacAddress = "AA:BB:CC:11:22:33"
	defaultBmHostBootMode   = "UEFISecureBoot"
	defaultBmHostAnnotation = "annotation.openshift.io/test-annotation"
	testSchemes             = []clients.SchemeAttacher{
		bmhv1alpha1.AddToScheme,
	}
)

func Test_BmhBuilder_Pull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull,
		bmhv1alpha1.AddToScheme,
		bmhv1alpha1.GroupVersion.WithKind("BareMetalHost"),
	).ExecuteTests(t)
}

func Test_BmhBuilder_Methods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[bmhv1alpha1.BareMetalHost, BmhBuilder](
		bmhv1alpha1.AddToScheme,
		bmhv1alpha1.GroupVersion.WithKind("BareMetalHost"),
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
		Run(t)
}

func TestBareMetalHostWithRootDeviceDeviceName(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedError    string
		deviceDeviceName string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:    "",
			deviceDeviceName: "123",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			deviceDeviceName: "",
			expectedError:    "the baremetalhost rootDeviceHint deviceName cannot be empty",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			deviceDeviceName: "123",
			expectedError:    "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceDeviceName(testCase.deviceDeviceName)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.deviceDeviceName, testBmHostBuilder.Definition.Spec.RootDeviceHints.DeviceName)
		}
	}
}

func TestBareMetalHostWithRootDeviceHTCL(t *testing.T) {
	testCases := []struct {
		testBmHost     *BmhBuilder
		expectedError  string
		rootDeviceHTCL string
	}{
		{
			testBmHost:     buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:  "",
			rootDeviceHTCL: "123",
		},
		{
			testBmHost:     buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceHTCL: "",
			expectedError:  "the baremetalhost rootDeviceHint hctl cannot be empty",
		},
		{
			testBmHost:     buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceHTCL: "123",
			expectedError:  "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceHTCL(testCase.rootDeviceHTCL)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.rootDeviceHTCL, testBmHostBuilder.Definition.Spec.RootDeviceHints.HCTL)
		}
	}
}

func TestBareMetalHostWithRootDeviceModel(t *testing.T) {
	testCases := []struct {
		testBmHost      *BmhBuilder
		expectedError   string
		rootDeviceModel string
	}{
		{
			testBmHost:      buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:   "",
			rootDeviceModel: "123",
		},
		{
			testBmHost:      buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceModel: "",
			expectedError:   "the baremetalhost rootDeviceHint model cannot be empty",
		},
		{
			testBmHost:      buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceModel: "123",
			expectedError:   "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceModel(testCase.rootDeviceModel)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.rootDeviceModel, testBmHostBuilder.Definition.Spec.RootDeviceHints.Model)
		}
	}
}

func TestBareMetalHostWithRootDeviceVendor(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedError    string
		rootDeviceVendor string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:    "",
			rootDeviceVendor: "123",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceVendor: "",
			expectedError:    "the baremetalhost rootDeviceHint vendor cannot be empty",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceVendor: "123",
			expectedError:    "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceVendor(testCase.rootDeviceVendor)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.rootDeviceVendor, testBmHostBuilder.Definition.Spec.RootDeviceHints.Model)
		}
	}
}

func TestBareMetalHostWithRootDeviceSerialNumber(t *testing.T) {
	testCases := []struct {
		testBmHost             *BmhBuilder
		expectedError          string
		rootDeviceSerialNumber string
	}{
		{
			testBmHost:             buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:          "",
			rootDeviceSerialNumber: "123",
		},
		{
			testBmHost:             buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceSerialNumber: "",
			expectedError:          "the baremetalhost rootDeviceHint serialNumber cannot be empty",
		},
		{
			testBmHost:             buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceSerialNumber: "123",
			expectedError:          "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceSerialNumber(testCase.rootDeviceSerialNumber)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceSerialNumber, testBmHostBuilder.Definition.Spec.RootDeviceHints.SerialNumber)
		}
	}
}

func TestBareMetalHostWithRootDeviceMinSizeGigabytes(t *testing.T) {
	testCases := []struct {
		testBmHost                 *BmhBuilder
		expectedError              string
		rootDeviceMinSizeGigabytes int
	}{
		{
			testBmHost:                 buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:              "",
			rootDeviceMinSizeGigabytes: 12,
		},
		{
			testBmHost:                 buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceMinSizeGigabytes: -1,
			expectedError:              "the baremetalhost rootDeviceHint size cannot be less than 0",
		},
		{
			testBmHost:                 buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceMinSizeGigabytes: 123,
			expectedError:              "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceMinSizeGigabytes(testCase.rootDeviceMinSizeGigabytes)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceMinSizeGigabytes, testBmHostBuilder.Definition.Spec.RootDeviceHints.MinSizeGigabytes)
		}
	}
}

func TestBareMetalHostWithRootDeviceWWN(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError string
		rootDeviceWwn string
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: "",
			rootDeviceWwn: "test",
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceWwn: "",
			expectedError: "the baremetalhost rootDeviceHint wwn cannot be empty",
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceWwn: "test",
			expectedError: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceWWN(testCase.rootDeviceWwn)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceWwn, testBmHostBuilder.Definition.Spec.RootDeviceHints.WWN)
		}
	}
}

func TestBareMetalHostWithRootDeviceWWNWithExtension(t *testing.T) {
	testCases := []struct {
		testBmHost                 *BmhBuilder
		expectedError              string
		rootDeviceWWNWithExtension string
	}{
		{
			testBmHost:                 buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:              "",
			rootDeviceWWNWithExtension: "test",
		},
		{
			testBmHost:                 buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceWWNWithExtension: "",
			expectedError:              "the baremetalhost rootDeviceHint wwnWithExtension cannot be empty",
		},
		{
			testBmHost:                 buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceWWNWithExtension: "test",
			expectedError:              "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceWWNWithExtension(testCase.rootDeviceWWNWithExtension)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceWWNWithExtension, testBmHostBuilder.Definition.Spec.RootDeviceHints.WWNWithExtension)
		}
	}
}

func TestBareMetalHostWithRootDeviceWWNVendorExtension(t *testing.T) {
	testCases := []struct {
		testBmHost                   *BmhBuilder
		expectedError                string
		rootDeviceWWNVendorExtension string
	}{
		{
			testBmHost:                   buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:                "",
			rootDeviceWWNVendorExtension: "test",
		},
		{
			testBmHost:                   buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceWWNVendorExtension: "",
			expectedError:                "the baremetalhost rootDeviceHint wwnVendorExtension cannot be empty",
		},
		{
			testBmHost:                   buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceWWNVendorExtension: "test",
			expectedError:                "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceWWNVendorExtension(testCase.rootDeviceWWNVendorExtension)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceWWNVendorExtension, testBmHostBuilder.Definition.Spec.RootDeviceHints.WWNVendorExtension)
		}
	}
}

func TestBareMetalHostWithRootDeviceRotationalDisk(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError string
		rotational    bool
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: "",
			rotational:    true,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
			rotational:    false,
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: "not acceptable 'bootMode' value",
			rotational:    false,
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceRotationalDisk(testCase.rotational)
		assert.Equal(t, testCase.expectedError, getErrorString(testBmHostBuilder))

		if testCase.expectedError == "" {
			assert.Equal(
				t, &testCase.rotational, testBmHostBuilder.Definition.Spec.RootDeviceHints.Rotational)
		}
	}
}

func TestBareMetalHostWithOptions(t *testing.T) {
	testSettings := buildBareMetalHostTestClientWithDummyObject()
	testBuilder := buildValidBmHostBuilder(testSettings).WithOptions(
		func(builder *BmhBuilder) (*BmhBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", getErrorString(testBuilder))
	testBuilder = buildValidBmHostBuilder(testSettings).WithOptions(
		func(builder *BmhBuilder) (*BmhBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", getErrorString(testBuilder))
}

func TestBareMetalHostGetBmhOperationalState(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedState bmhv1alpha1.OperationalStatus
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedState: bmhv1alpha1.OperationalStatusOK,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedState: "",
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedState: "",
		},
	}

	for _, testCase := range testCases {
		bmhOperationalState := testCase.testBmHost.GetBmhOperationalState()
		assert.Equal(t, testCase.expectedState, bmhOperationalState)
	}
}

func TestBareMetalHostGetBmhPowerOnStatus(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedState bool
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedState: true,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedState: false,
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedState: false,
		},
	}

	for _, testCase := range testCases {
		bmhPowerOnStatus := testCase.testBmHost.GetBmhPowerOnStatus()
		assert.Equal(t, testCase.expectedState, bmhPowerOnStatus)
	}
}

func TestBareMetalHostCreateAndWaitUntilProvisioned(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedErrorMsg string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		ipAddressPoolBuilder, err := testCase.testBmHost.CreateAndWaitUntilProvisioned(1 * time.Millisecond)
		if testCase.expectedErrorMsg == "" {
			assert.Nil(t, err)
			assert.Equal(t, ipAddressPoolBuilder.Definition.Name, ipAddressPoolBuilder.Object.Name)
			assert.Equal(t, ipAddressPoolBuilder.Definition.Namespace, ipAddressPoolBuilder.Object.Namespace)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		}
	}
}

func TestBareMetalHostWaitUntilProvisioned(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedErrorMsg string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "",
		},
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateDeprovisioning)),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilProvisioned(1 * time.Millisecond)
		if testCase.expectedErrorMsg != "" {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostWaitUntilProvisioning(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedErrorMsg string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateProvisioning)),
			expectedErrorMsg: "",
		},
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilProvisioning(1 * time.Millisecond)
		if testCase.expectedErrorMsg != "" {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostWaitUntilReady(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedErrorMsg string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateReady)),
			expectedErrorMsg: "",
		},
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilReady(1 * time.Millisecond)
		if testCase.expectedErrorMsg != "" {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostWaitUntilAvailable(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedErrorMsg string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateAvailable)),
			expectedErrorMsg: "",
		},
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilAvailable(1 * time.Millisecond)
		if testCase.expectedErrorMsg != "" {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostWaitUntilInStatus(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedErrorMsg string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateAvailable)),
			expectedErrorMsg: "",
		},
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateProvisioning)),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilInStatus(bmhv1alpha1.StateAvailable, 1*time.Millisecond)
		if testCase.expectedErrorMsg != "" {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostDeleteAndWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedErrorMsg string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorMsg: "",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		builder, err := testCase.testBmHost.DeleteAndWaitUntilDeleted(2 * time.Second)

		if testCase.expectedErrorMsg == "" {
			assert.Nil(t, err)
			assert.Nil(t, testCase.testBmHost.Object)
			assert.Nil(t, builder)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		}
	}
}

func TestBareMetalHostWaitUntilAnnotationExists(t *testing.T) {
	testCases := []struct {
		annotation       string
		exists           bool
		valid            bool
		annotated        bool
		expectedErrorMsg string
	}{
		{
			annotation:       defaultBmHostAnnotation,
			exists:           true,
			valid:            true,
			annotated:        true,
			expectedErrorMsg: "",
		},
		{
			annotation:       "",
			exists:           true,
			valid:            true,
			annotated:        true,
			expectedErrorMsg: "bmh annotation key cannot be empty",
		},
		{
			annotation:       defaultBmHostAnnotation,
			exists:           false,
			valid:            true,
			annotated:        true,
			expectedErrorMsg: "does not exist",
		},
		{
			annotation:       defaultBmHostAnnotation,
			exists:           true,
			valid:            false,
			annotated:        true,
			expectedErrorMsg: "not acceptable 'bootMode' value",
		},
		{
			annotation:       defaultBmHostAnnotation,
			exists:           true,
			valid:            true,
			annotated:        false,
			expectedErrorMsg: "context deadline exceeded",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			builder        *BmhBuilder
		)

		if testCase.exists {
			bmh := buildDummyBmHost(bmhv1alpha1.StateProvisioned, bmhv1alpha1.OperationalStatusOK)

			if testCase.annotated {
				bmh.Annotations = map[string]string{
					defaultBmHostAnnotation: "",
				}
			}

			runtimeObjects = append(runtimeObjects, bmh)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		if testCase.valid {
			builder = buildValidBmHostBuilder(testSettings)
		} else {
			builder = buildInValidBmHostBuilder(testSettings)
		}

		_, err := builder.WaitUntilAnnotationExists(testCase.annotation, time.Second)

		if testCase.expectedErrorMsg == "" {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		}
	}
}

func TestBareMetalHostWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedErrorMsg string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorMsg: "",
		},
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "context deadline exceeded",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedErrorMsg: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilDeleted(2 * time.Second)

		if testCase.expectedErrorMsg == "" {
			assert.Nil(t, err)
			assert.Nil(t, testCase.testBmHost.Object)
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		}
	}
}

func buildValidBmHostBuilder(apiClient *clients.Settings) *BmhBuilder {
	return NewBuilder(
		apiClient,
		defaultBmHostName,
		defaultBmHostNsName,
		defaultBmHostAddress,
		defaultBmHostSecretName,
		defaultBmHostMacAddress,
		defaultBmHostBootMode,
	)
}

func buildInValidBmHostBuilder(apiClient *clients.Settings) *BmhBuilder {
	return NewBuilder(
		apiClient,
		defaultBmHostName,
		defaultBmHostNsName,
		defaultBmHostAddress,
		defaultBmHostSecretName,
		defaultBmHostMacAddress,
		"test",
	)
}

func buildBareMetalHostTestClientWithDummyObject(state ...bmhv1alpha1.ProvisioningState) *clients.Settings {
	provisionState := bmhv1alpha1.StateProvisioned
	if len(state) > 0 {
		provisionState = state[0]
	}

	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyBmHostObject(provisionState),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyBmHost(
	state bmhv1alpha1.ProvisioningState, operState bmhv1alpha1.OperationalStatus) *bmhv1alpha1.BareMetalHost {
	return &bmhv1alpha1.BareMetalHost{
		Spec: bmhv1alpha1.BareMetalHostSpec{
			BMC: bmhv1alpha1.BMCDetails{
				Address:                        defaultBmHostAddress,
				CredentialsName:                defaultBmHostSecretName,
				DisableCertificateVerification: true,
			},
			BootMode:              bmhv1alpha1.BootMode(defaultBmHostBootMode),
			BootMACAddress:        defaultBmHostMacAddress,
			Online:                true,
			ExternallyProvisioned: false,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultBmHostName,
			Namespace: defaultBmHostNsName,
		},
		Status: bmhv1alpha1.BareMetalHostStatus{
			OperationalStatus: operState,
			PoweredOn:         true,
			Provisioning: bmhv1alpha1.ProvisionStatus{
				State: state,
			},
		},
	}
}

func buildDummyBmHostObject(
	state bmhv1alpha1.ProvisioningState, operationalStatus ...bmhv1alpha1.OperationalStatus) []runtime.Object {
	operState := bmhv1alpha1.OperationalStatusOK
	if len(operationalStatus) > 0 {
		operState = operationalStatus[0]
	}

	return append([]runtime.Object{}, buildDummyBmHost(state, operState))
}

// getErrorString returns the error string from the builder, or empty string if no error.
func getErrorString(builder *BmhBuilder) string {
	if builder == nil {
		return ""
	}

	err := builder.GetError()
	if err == nil {
		return ""
	}

	return err.Error()
}
