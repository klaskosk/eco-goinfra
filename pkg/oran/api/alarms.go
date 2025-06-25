package api

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/filter"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/internal/alarms"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/internal/common"
	"k8s.io/utils/ptr"
)

// AlarmEventRecord is the type of the AlarmEventRecord resource returned by the API.
type AlarmEventRecord = alarms.AlarmEventRecord

// AlarmEventRecordModifications represents modifications that can be applied to an alarm event record.
type AlarmEventRecordModifications = alarms.AlarmEventRecordModifications

// AlarmServiceConfiguration represents the configuration for the alarm service.
type AlarmServiceConfiguration = alarms.AlarmServiceConfiguration

// AlarmSubscriptionInfo represents information about an alarm subscription.
type AlarmSubscriptionInfo = alarms.AlarmSubscriptionInfo

// AlarmSubscriptionInfoFilter represents filter criteria for alarm subscriptions.
type AlarmSubscriptionInfoFilter = alarms.AlarmSubscriptionInfoFilter

// PerceivedSeverity represents the perceived severity of an alarm.
type PerceivedSeverity = alarms.PerceivedSeverity

// AlarmsClient is a client for the O-RAN O2IMS Alarms API.
type AlarmsClient struct {
	alarms.ClientWithResponsesInterface
}

// ListAlarms lists all alarm event records. Optionally, a filter can be provided
// to filter the list of alarms. If more than one filter is provided, only the first one is used. filter.And() can be
// used to combine multiple filters.
func (client *AlarmsClient) ListAlarms(filter ...filter.Filter) ([]AlarmEventRecord, error) {
	var filterString *common.Filter

	if len(filter) > 0 {
		filterString = ptr.To(filter[0].Filter())
	}

	resp, err := client.GetAlarmsWithResponse(context.TODO(),
		&alarms.GetAlarmsParams{Filter: filterString})
	if err != nil {
		return nil, fmt.Errorf("failed to list AlarmEventRecords: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list AlarmEventRecords: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// GetAlarm gets an alarm event record by its ID.
func (client *AlarmsClient) GetAlarm(id string) (*AlarmEventRecord, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get AlarmEventRecord: invalid UUID format: %w", err)
	}

	resp, err := client.GetAlarmWithResponse(context.TODO(), parsedUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AlarmEventRecord: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get AlarmEventRecord: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// PatchAlarm modifies an alarm event record by its ID.
func (client *AlarmsClient) PatchAlarm(
	id string, modifications AlarmEventRecordModifications) (*AlarmEventRecordModifications, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to patch AlarmEventRecord: invalid UUID format: %w", err)
	}

	resp, err := client.PatchAlarmWithApplicationMergePatchPlusJSONBodyWithResponse(
		context.TODO(), parsedUUID, modifications)
	if err != nil {
		return nil, fmt.Errorf("failed to patch AlarmEventRecord: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to patch AlarmEventRecord: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// GetServiceConfiguration gets the alarm service configuration.
func (client *AlarmsClient) GetServiceConfiguration() (*AlarmServiceConfiguration, error) {
	resp, err := client.GetServiceConfigurationWithResponse(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get AlarmServiceConfiguration: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get AlarmServiceConfiguration: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// UpdateServiceConfiguration updates the alarm service configuration.
func (client *AlarmsClient) UpdateServiceConfiguration(config AlarmServiceConfiguration) (
	*AlarmServiceConfiguration, error) {
	resp, err := client.UpdateAlarmServiceConfigurationWithResponse(context.TODO(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to update AlarmServiceConfiguration: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to update AlarmServiceConfiguration: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// PatchServiceConfiguration partially updates the alarm service configuration.
func (client *AlarmsClient) PatchServiceConfiguration(config AlarmServiceConfiguration) (
	*AlarmServiceConfiguration, error) {
	resp, err := client.PatchAlarmServiceConfigurationWithApplicationMergePatchPlusJSONBodyWithResponse(
		context.TODO(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to patch AlarmServiceConfiguration: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to patch AlarmServiceConfiguration: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// ListSubscriptions lists all alarm subscriptions. Optionally, a filter can be provided
// to filter the list of subscriptions. If more than one filter is provided, only the first one is used.
// filter.And() can be used to combine multiple filters.
func (client *AlarmsClient) ListSubscriptions(filter ...filter.Filter) ([]AlarmSubscriptionInfo, error) {
	var filterString *common.Filter

	if len(filter) > 0 {
		filterString = ptr.To(filter[0].Filter())
	}

	resp, err := client.GetSubscriptionsWithResponse(context.TODO(),
		&alarms.GetSubscriptionsParams{Filter: filterString})
	if err != nil {
		return nil, fmt.Errorf("failed to list AlarmSubscriptionInfos: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list AlarmSubscriptionInfos: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// CreateSubscription creates a new alarm subscription.
func (client *AlarmsClient) CreateSubscription(subscription AlarmSubscriptionInfo) (*AlarmSubscriptionInfo, error) {
	// Validate callback URL
	if _, err := url.ParseRequestURI(subscription.Callback); err != nil {
		return nil, fmt.Errorf("failed to create AlarmSubscriptionInfo: invalid callback URL: %w", err)
	}

	resp, err := client.CreateSubscriptionWithResponse(context.TODO(), subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to create AlarmSubscriptionInfo: error contacting api: %w", err)
	}

	if resp.StatusCode() != 201 || resp.JSON201 == nil {
		return nil, fmt.Errorf("failed to create AlarmSubscriptionInfo: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return resp.JSON201, nil
}

// GetSubscription gets an alarm subscription by its ID.
func (client *AlarmsClient) GetSubscription(id string) (*AlarmSubscriptionInfo, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get AlarmSubscriptionInfo: invalid UUID format: %w", err)
	}

	resp, err := client.GetSubscriptionWithResponse(context.TODO(), parsedUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AlarmSubscriptionInfo: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get AlarmSubscriptionInfo: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// DeleteSubscription deletes an alarm subscription by its ID.
func (client *AlarmsClient) DeleteSubscription(id string) error {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("failed to delete AlarmSubscriptionInfo: invalid UUID format: %w", err)
	}

	resp, err := client.DeleteSubscriptionWithResponse(context.TODO(), parsedUUID)
	if err != nil {
		return fmt.Errorf("failed to delete AlarmSubscriptionInfo: error contacting api: %w", err)
	}

	if resp.StatusCode() != 204 {
		return fmt.Errorf("failed to delete AlarmSubscriptionInfo: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return nil
}
