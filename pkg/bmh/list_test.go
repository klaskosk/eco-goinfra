package bmh

import (
	"context"
	"fmt"
	"testing"
	"time"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBareMetalHostList(t *testing.T) {
	testCases := []struct {
		BareMetalHosts   []*BmhBuilder
		nsName           string
		listOptions      []goclient.ListOption
		expectedErrorMsg string
		client           bool
	}{
		{

			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			expectedErrorMsg: "",
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "",
			expectedErrorMsg: "nsname",
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			listOptions:      []goclient.ListOption{goclient.MatchingLabelsSelector{Selector: labels.NewSelector()}},
			client:           true,
			expectedErrorMsg: "",
		},
		{
			BareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:         "test-namespace",
			listOptions: []goclient.ListOption{
				goclient.MatchingLabelsSelector{Selector: labels.NewSelector()},
				goclient.MatchingLabelsSelector{Selector: labels.NewSelector()},
			},
			expectedErrorMsg: "more than one ListOptions",
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			expectedErrorMsg: "apiClient",
			client:           false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyBmHostObject(bmhv1alpha1.StateProvisioned),
				SchemeAttachers: testSchemes,
			})
		}

		bmhBuilders, err := List(testSettings, testCase.nsName, testCase.listOptions...)

		if testCase.expectedErrorMsg == "" {
			assert.Nil(t, err)

			if len(testCase.listOptions) == 0 {
				assert.Equal(t, len(testCase.BareMetalHosts), len(bmhBuilders))
			}
		} else {
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), testCase.expectedErrorMsg)
		}
	}
}

func TestBareMetalHostListInAllNamespaces(t *testing.T) {
	testCases := []struct {
		bareMetalHosts []*BmhBuilder
		listOptions    []goclient.ListOption
		expectedError  error
		client         bool
	}{
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    nil,
			expectedError:  nil,
			client:         true,
		},
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    []goclient.ListOption{goclient.MatchingLabelsSelector{Selector: labels.NewSelector()}},
			expectedError:  nil,
			client:         true,
		},
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions: []goclient.ListOption{
				goclient.MatchingLabelsSelector{Selector: labels.NewSelector()},
				goclient.MatchingLabelsSelector{Selector: labels.NewSelector()},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			bareMetalHosts: []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			listOptions:    []goclient.ListOption{goclient.MatchingLabelsSelector{Selector: labels.NewSelector()}},
			expectedError:  fmt.Errorf("failed to list bareMetalHosts, 'apiClient' parameter is empty"),
			client:         false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildBareMetalHostTestClientWithDummyObject()
		}

		bmhBuilders, err := ListInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.bareMetalHosts), len(bmhBuilders))
		}
	}
}

func TestBareMetalWaitForAllBareMetalHostsInGoodOperationalState(t *testing.T) {
	testCases := []struct {
		BareMetalHosts   []*BmhBuilder
		nsName           string
		listOptions      []goclient.ListOption
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
			expectedError:    fmt.Errorf("failed to list bareMetalHosts, 'nsname' parameter is empty"),
			expectedStatus:   false,
			listOptions:      nil,
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    fmt.Errorf("failed to list bareMetalHosts, 'apiClient' parameter is empty"),
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
			listOptions:      []goclient.ListOption{goclient.MatchingLabelsSelector{Selector: labels.NewSelector()}},
			client:           true,
		},
		{
			BareMetalHosts:   []*BmhBuilder{buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject())},
			nsName:           "test-namespace",
			operationalState: bmhv1alpha1.OperationalStatusOK,
			expectedError:    fmt.Errorf("error: more than one ListOptions was passed"),
			expectedStatus:   false,
			listOptions: []goclient.ListOption{
				goclient.MatchingLabelsSelector{Selector: labels.NewSelector()}, goclient.MatchingLabelsSelector{Selector: labels.NewSelector()}},
			client: true,
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
