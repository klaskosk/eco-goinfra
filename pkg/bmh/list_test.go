package bmh

import (
	"context"
	"testing"
	"time"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_BmhBuilder_List(t *testing.T) {
	testhelper.NewNamespacedListTestConfig(
		List,
		bmhv1alpha1.AddToScheme,
		bmhv1alpha1.GroupVersion.WithKind("BareMetalHost"),
	).ExecuteTests(t)
}

func Test_BmhBuilder_ListInAllNamespaces(t *testing.T) {
	testhelper.NewListTestConfig(
		ListInAllNamespaces,
		bmhv1alpha1.AddToScheme,
		bmhv1alpha1.GroupVersion.WithKind("BareMetalHost"),
	).ExecuteTests(t)
}

func TestBareMetalWaitForAllBareMetalHostsInGoodOperationalState(t *testing.T) {
	testCases := []struct {
		BareMetalHosts   []*BmhBuilder
		nsName           string
		listOptions      []goclient.ListOptions
		operationalState bmhv1alpha1.OperationalStatus
		expectedError    error
		client           bool
		expectedStatus   bool
	}{
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			listOptions:      nil,
			expectedError:    nil,
			expectedStatus:   true,
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusDelayed,
			expectedError:    context.DeadlineExceeded,
			listOptions:      nil,
			expectedStatus:   false,
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError: commonerrors.NewBuilderFieldEmpty(
				key.NewResourceKey("BareMetalHost", "", ""), commonerrors.BuilderFieldNamespace),
			expectedStatus: false,
			listOptions:    nil,
			client:         true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    commonerrors.NewAPIClientNil(key.NewResourceKey("BareMetalHost", "", "")),
			expectedStatus:   false,
			listOptions:      nil,
			client:           false,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    nil,
			expectedStatus:   true,
			listOptions:      []goclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:           true,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyBmHostObject(bmhv1alpha1.StateProvisioned, testCase.operationalState),
				SchemeAttachers: testSchemes,
			})
		}

		status, err := WaitForAllBareMetalHostsInGoodOperationalState(
			testSettings, testCase.nsName, 1*time.Second, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedStatus, status)
	}
}
