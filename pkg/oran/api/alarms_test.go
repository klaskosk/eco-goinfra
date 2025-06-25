package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/filter"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/internal/alarms"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

var (
	// dummyAlarmEventRecord is a test alarm event record for use in tests.
	dummyAlarmEventRecord = alarms.AlarmEventRecord{
		AlarmEventRecordId: uuid.New(),
		AlarmDefinitionID:  uuid.New(),
		ResourceTypeID:     uuid.New(),
		ResourceID:         uuid.New(),
		ProbableCauseID:    uuid.New(),
		PerceivedSeverity:  alarms.CRITICAL,
		AlarmRaisedTime:    time.Now(),
	}

	// dummyAlarmEventRecordModifications is a test alarm event record modifications object for use in tests.
	dummyAlarmEventRecordModifications = alarms.AlarmEventRecordModifications{
		AlarmAcknowledged: ptr.To(true),
	}

	// dummyAlarmServiceConfiguration is a test alarm service configuration for use in tests.
	dummyAlarmServiceConfiguration = alarms.AlarmServiceConfiguration{
		RetentionPeriod: 10,
		Extensions:      &map[string]string{"key1": "value1"},
	}

	// dummyAlarmSubscriptionInfo is a test alarm subscription for use in tests.
	dummyAlarmSubscriptionInfo = alarms.AlarmSubscriptionInfo{
		AlarmSubscriptionId:    ptr.To(uuid.New()),
		Callback:               "http://test.com/callback",
		ConsumerSubscriptionId: ptr.To(uuid.New()),
	}

	// testAlarmID is a test alarm ID for use in tests.
	testAlarmID = dummyAlarmEventRecord.AlarmEventRecordId.String()

	// testSubscriptionID is a test subscription ID for use in tests.
	testSubscriptionID = dummyAlarmSubscriptionInfo.AlarmSubscriptionId.String()
)

func TestAlarmsListAlarms(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		filter        []filter.Filter
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success without filter",
			filter:  nil,
			handler: successHandler([]alarms.AlarmEventRecord{dummyAlarmEventRecord}, http.StatusOK),
		},
		{
			name:    "success with filter",
			filter:  []filter.Filter{filter.Equals("name", "test-alarm")},
			handler: filterSuccessHandler([]alarms.AlarmEventRecord{dummyAlarmEventRecord}, "(eq,name,test-alarm)"),
		},
		{
			name:    "success with multiple filters - only first used",
			filter:  []filter.Filter{filter.Equals("name", "test1"), filter.Equals("version", "v1.0")},
			handler: filterSuccessHandler([]alarms.AlarmEventRecord{dummyAlarmEventRecord}, "(eq,name,test1)"),
		},
		{
			name:          "server error 500",
			filter:        nil,
			handler:       problemDetailsHandler(dummyProblemDetails),
			expectedError: "failed to list AlarmEventRecords: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.ListAlarms(testCase.filter...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAlarmsGetAlarm(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		alarmID       string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			alarmID: testAlarmID,
			handler: successHandler(dummyAlarmEventRecord, http.StatusOK),
		},
		{
			name:          "invalid uuid",
			alarmID:       "invalid-uuid",
			handler:       successHandler(dummyAlarmEventRecord, http.StatusOK),
			expectedError: "failed to get AlarmEventRecord: invalid UUID format:",
		},
		{
			name:          "server error 500",
			alarmID:       testAlarmID,
			handler:       problemDetailsHandler(dummyProblemDetails),
			expectedError: "failed to get AlarmEventRecord: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.GetAlarm(testCase.alarmID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAlarmsPatchAlarm(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		alarmID       string
		modifications AlarmEventRecordModifications
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:          "success",
			alarmID:       testAlarmID,
			modifications: dummyAlarmEventRecordModifications,
			handler:       successHandler(dummyAlarmEventRecordModifications, http.StatusOK),
		},
		{
			name:          "invalid uuid",
			alarmID:       "invalid-uuid",
			modifications: dummyAlarmEventRecordModifications,
			handler:       successHandler(dummyAlarmEventRecordModifications, http.StatusOK),
			expectedError: "failed to patch AlarmEventRecord: invalid UUID format:",
		},
		{
			name:          "server error 500",
			alarmID:       testAlarmID,
			modifications: dummyAlarmEventRecordModifications,
			handler:       problemDetailsHandler(dummyProblemDetails),
			expectedError: "failed to patch AlarmEventRecord: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.PatchAlarm(testCase.alarmID, testCase.modifications)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAlarmsGetServiceConfiguration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			handler: successHandler(dummyAlarmServiceConfiguration, http.StatusOK),
		},
		{
			name:          "server error 500",
			handler:       problemDetailsHandler(dummyProblemDetails),
			expectedError: "failed to get AlarmServiceConfiguration: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.GetServiceConfiguration()

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAlarmsUpdateServiceConfiguration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		config        AlarmServiceConfiguration
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			config:  dummyAlarmServiceConfiguration,
			handler: successHandler(dummyAlarmServiceConfiguration, http.StatusOK),
		},
		{
			name:          "server error 500",
			config:        dummyAlarmServiceConfiguration,
			handler:       problemDetailsHandler(dummyProblemDetails),
			expectedError: "failed to update AlarmServiceConfiguration: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.UpdateServiceConfiguration(testCase.config)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAlarmsPatchServiceConfiguration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		config        AlarmServiceConfiguration
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success",
			config:  dummyAlarmServiceConfiguration,
			handler: successHandler(dummyAlarmServiceConfiguration, http.StatusOK),
		},
		{
			name:          "server error 500",
			config:        dummyAlarmServiceConfiguration,
			handler:       problemDetailsHandler(dummyProblemDetails),
			expectedError: "failed to patch AlarmServiceConfiguration: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.PatchServiceConfiguration(testCase.config)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAlarmsListSubscriptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		filter        []filter.Filter
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:    "success without filter",
			filter:  nil,
			handler: successHandler([]alarms.AlarmSubscriptionInfo{dummyAlarmSubscriptionInfo}, http.StatusOK),
		},
		{
			name:   "success with filter",
			filter: []filter.Filter{filter.Equals("callback", "http://test.com/callback")},
			handler: filterSuccessHandler([]alarms.AlarmSubscriptionInfo{dummyAlarmSubscriptionInfo},
				"(eq,callback,http://test.com/callback)"),
		},
		{
			name:   "success with multiple filters - only first used",
			filter: []filter.Filter{filter.Equals("callback", "http://first.com"), filter.Equals("status", "active")},
			handler: filterSuccessHandler([]alarms.AlarmSubscriptionInfo{dummyAlarmSubscriptionInfo},
				"(eq,callback,http://first.com)"),
		},
		{
			name:          "server error 500",
			filter:        nil,
			handler:       problemDetailsHandler(dummyProblemDetails),
			expectedError: "failed to list AlarmSubscriptionInfos: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.ListSubscriptions(testCase.filter...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAlarmsCreateSubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		subscription  AlarmSubscriptionInfo
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:         "success",
			subscription: dummyAlarmSubscriptionInfo,
			handler:      successHandler(dummyAlarmSubscriptionInfo, http.StatusCreated),
		},
		{
			name: "invalid callback url",
			subscription: alarms.AlarmSubscriptionInfo{
				Callback: "invalid-url",
			},
			handler:       successHandler(dummyAlarmSubscriptionInfo, http.StatusCreated),
			expectedError: "failed to create AlarmSubscriptionInfo: invalid callback URL:",
		},
		{
			name:          "server error 500",
			subscription:  dummyAlarmSubscriptionInfo,
			handler:       problemDetailsHandler(dummyProblemDetails),
			expectedError: "failed to create AlarmSubscriptionInfo: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.CreateSubscription(testCase.subscription)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAlarmsGetSubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID string
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: testSubscriptionID,
			handler:        successHandler(dummyAlarmSubscriptionInfo, http.StatusOK),
		},
		{
			name:           "invalid uuid",
			subscriptionID: "invalid-uuid",
			handler:        successHandler(dummyAlarmSubscriptionInfo, http.StatusOK),
			expectedError:  "failed to get AlarmSubscriptionInfo: invalid UUID format:",
		},
		{
			name:           "server error 500",
			subscriptionID: testSubscriptionID,
			handler:        problemDetailsHandler(dummyProblemDetails),
			expectedError:  "failed to get AlarmSubscriptionInfo: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			result, err := alarmsClient.GetSubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestAlarmsDeleteSubscription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		subscriptionID string
		handler        http.HandlerFunc
		expectedError  string
	}{
		{
			name:           "success",
			subscriptionID: testSubscriptionID,
			handler:        successHandler(nil, http.StatusNoContent),
		},
		{
			name:           "invalid uuid",
			subscriptionID: "invalid-uuid",
			handler:        successHandler(nil, http.StatusNoContent),
			expectedError:  "failed to delete AlarmSubscriptionInfo: invalid UUID format:",
		},
		{
			name:           "server error 500",
			subscriptionID: testSubscriptionID,
			handler:        problemDetailsHandler(dummyProblemDetails),
			expectedError:  "failed to delete AlarmSubscriptionInfo: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := alarms.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			err = alarmsClient.DeleteSubscription(testCase.subscriptionID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

//nolint:funlen // This is only long because it must test all the functions.
func TestAlarmsNetworkError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		testFunc func(client *AlarmsClient) error
	}{
		{
			name: "ListAlarms network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.ListAlarms()

				return err
			},
		},
		{
			name: "GetAlarm network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.GetAlarm(uuid.NewString())

				return err
			},
		},
		{
			name: "PatchAlarm network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.PatchAlarm(uuid.NewString(), dummyAlarmEventRecordModifications)

				return err
			},
		},
		{
			name: "GetServiceConfiguration network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.GetServiceConfiguration()

				return err
			},
		},
		{
			name: "UpdateServiceConfiguration network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.UpdateServiceConfiguration(dummyAlarmServiceConfiguration)

				return err
			},
		},
		{
			name: "PatchServiceConfiguration network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.PatchServiceConfiguration(dummyAlarmServiceConfiguration)

				return err
			},
		},
		{
			name: "ListSubscriptions network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.ListSubscriptions()

				return err
			},
		},
		{
			name: "CreateSubscription network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.CreateSubscription(dummyAlarmSubscriptionInfo)

				return err
			},
		},
		{
			name: "GetSubscription network error",
			testFunc: func(client *AlarmsClient) error {
				_, err := client.GetSubscription(uuid.NewString())

				return err
			},
		},
		{
			name: "DeleteSubscription network error",
			testFunc: func(client *AlarmsClient) error {
				err := client.DeleteSubscription(uuid.NewString())

				return err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// 192.0.2.0 is a reserved test address so we never accidentally use a valid IP. Still, we set a
			// timeout to ensure that we do not timeout the test.
			client, err := alarms.NewClientWithResponses("http://192.0.2.0:8080",
				alarms.WithHTTPClient(&http.Client{Timeout: time.Second * 1}))
			assert.NoError(t, err)

			alarmsClient := &AlarmsClient{ClientWithResponsesInterface: client}
			err = testCase.testFunc(alarmsClient)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error contacting api")
		})
	}
}
