package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/filter"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/inventory"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

// OCloudInfo is the type of the OCloudInfo resource returned by the API.
type OCloudInfo = inventory.OCloudInfo

// LocationInfo is the type of the LocationInfo resource returned by the API.
type LocationInfo = inventory.LocationInfo

// OCloudSiteInfo is the type of the OCloudSiteInfo resource returned by the API.
type OCloudSiteInfo = inventory.OCloudSiteInfo

// ResourcePool is the type of the ResourcePool resource returned by the API.
type ResourcePool = inventory.ResourcePool

// Resource is the type of the Resource resource returned by the API.
type Resource = inventory.Resource

// ResourceType is the type of the ResourceType resource returned by the API.
type ResourceType = inventory.ResourceType

// AlarmDictionary is the type of the AlarmDictionary resource returned by the API.
type AlarmDictionary = common.AlarmDictionary

// DeploymentManager is the type of the DeploymentManager resource returned by the API.
type DeploymentManager = inventory.DeploymentManager

// InventorySubscription is the type of the Subscription resource returned by the API.
type InventorySubscription = inventory.Subscription

// InventoryChangeNotification is the type of the InventoryChangeNotification resource returned by the API.
type InventoryChangeNotification = inventory.InventoryChangeNotification

// InventoryChangeNotificationType is the type of the notificationEventType field returned by the API.
type InventoryChangeNotificationType = inventory.InventoryChangeNotificationNotificationEventType

//nolint:revive // These are just re-exported constants no need for the linting.
const (
	InventoryChangeNotificationTypeCreate InventoryChangeNotificationType = inventory.N0
	InventoryChangeNotificationTypeModify InventoryChangeNotificationType = inventory.N1
	InventoryChangeNotificationTypeDelete InventoryChangeNotificationType = inventory.N2
)

// APIVersions is the type of the APIVersions resource returned by the API.
type APIVersions = common.APIVersions

// APIVersion is the type of the APIVersion resource returned by the API.
type APIVersion = common.APIVersion

// InventoryListParams contains optional query parameters for inventory list endpoints.
type InventoryListParams struct {
	Fields         *string
	ExcludeFields  *string
	Filter         filter.Filter
}

func (params *InventoryListParams) deploymentManagersParams() *inventory.GetDeploymentManagersParams {
	if params == nil {
		return &inventory.GetDeploymentManagersParams{}
	}

	return &inventory.GetDeploymentManagersParams{
		Fields:        (*common.Fields)(params.Fields),
		ExcludeFields: (*common.ExcludeFields)(params.ExcludeFields),
		Filter:        filterToCommon(params.Filter),
	}
}

func (params *InventoryListParams) locationsParams() *inventory.GetLocationsParams {
	if params == nil {
		return &inventory.GetLocationsParams{}
	}

	return &inventory.GetLocationsParams{
		Fields:        (*common.Fields)(params.Fields),
		ExcludeFields: (*common.ExcludeFields)(params.ExcludeFields),
		Filter:        filterToCommon(params.Filter),
	}
}

func filterToCommon(f filter.Filter) *common.Filter {
	if f == nil {
		return nil
	}

	return ptr.To(f.Filter())
}

// InventoryClient provides access to the O2IMS infrastructure inventory API.
type InventoryClient struct {
	inventory.ClientWithResponsesInterface
}

// GetAPIVersions retrieves the complete list of API versions implemented by the service.
func (client *InventoryClient) GetAPIVersions() (APIVersions, error) {
	resp, err := client.GetAllVersionsWithResponse(context.TODO())
	if err != nil {
		return APIVersions{}, fmt.Errorf("failed to get API versions: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return APIVersions{}, fmt.Errorf("failed to get API versions: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// GetCloudInfo retrieves O-Cloud instance metadata.
func (client *InventoryClient) GetCloudInfo() (*OCloudInfo, error) {
	resp, err := client.GetCloudInfoWithResponse(context.TODO(), &inventory.GetCloudInfoParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get O-Cloud info: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get O-Cloud info: received error from api: %w", apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// GetMinorAPIVersions retrieves the list of minor API versions for major version 1.
func (client *InventoryClient) GetMinorAPIVersions() (APIVersions, error) {
	resp, err := client.GetMinorVersionsWithResponse(context.TODO())
	if err != nil {
		return APIVersions{}, fmt.Errorf("failed to get minor API versions: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return APIVersions{}, fmt.Errorf("failed to get minor API versions: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// ListLocations lists all locations. Optionally, list parameters can be provided.
func (client *InventoryClient) ListLocations(params ...*InventoryListParams) ([]LocationInfo, error) {
	var listParams *InventoryListParams
	if len(params) > 0 {
		listParams = params[0]
	}

	resp, err := client.GetLocationsWithResponse(context.TODO(), listParams.locationsParams())
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list locations: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// GetLocation retrieves a location by its globalLocationId.
func (client *InventoryClient) GetLocation(globalLocationID string) (*LocationInfo, error) {
	resp, err := client.GetLocationWithResponse(context.TODO(), globalLocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get location: received error from api: %w", apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// ListOCloudSites lists all O-Cloud sites.
func (client *InventoryClient) ListOCloudSites() ([]OCloudSiteInfo, error) {
	resp, err := client.GetOCloudSitesWithResponse(context.TODO(), &inventory.GetOCloudSitesParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list O-Cloud sites: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list O-Cloud sites: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// GetOCloudSite retrieves an O-Cloud site by its oCloudSiteId.
func (client *InventoryClient) GetOCloudSite(oCloudSiteID uuid.UUID) (*OCloudSiteInfo, error) {
	resp, err := client.GetOCloudSiteWithResponse(context.TODO(), oCloudSiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get O-Cloud site: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get O-Cloud site: received error from api: %w", apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// ListResourcePools lists all resource pools.
func (client *InventoryClient) ListResourcePools() ([]ResourcePool, error) {
	resp, err := client.GetResourcePoolsWithResponse(context.TODO(), &inventory.GetResourcePoolsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list resource pools: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list resource pools: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// GetResourcePool retrieves a resource pool by its resourcePoolId.
func (client *InventoryClient) GetResourcePool(resourcePoolID uuid.UUID) (*ResourcePool, error) {
	resp, err := client.GetResourcePoolWithResponse(context.TODO(), resourcePoolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource pool: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get resource pool: received error from api: %w", apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// ListResources lists all resources in a resource pool.
func (client *InventoryClient) ListResources(resourcePoolID uuid.UUID) ([]Resource, error) {
	resp, err := client.GetResourcesWithResponse(context.TODO(), resourcePoolID, &inventory.GetResourcesParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list resources: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// GetResource retrieves a resource by its resourcePoolId and resourceId.
func (client *InventoryClient) GetResource(resourcePoolID, resourceID uuid.UUID) (*Resource, error) {
	resp, err := client.GetResourceWithResponse(context.TODO(), resourcePoolID, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get resource: received error from api: %w", apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// ListResourceTypes lists all resource types.
func (client *InventoryClient) ListResourceTypes() ([]ResourceType, error) {
	resp, err := client.GetResourceTypesWithResponse(context.TODO(), &inventory.GetResourceTypesParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list resource types: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list resource types: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// GetResourceType retrieves a resource type by its resourceTypeId.
func (client *InventoryClient) GetResourceType(resourceTypeID uuid.UUID) (*ResourceType, error) {
	resp, err := client.GetResourceTypeWithResponse(context.TODO(), resourceTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource type: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get resource type: received error from api: %w", apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// GetResourceTypeAlarmDictionary retrieves the alarm dictionary for a resource type.
func (client *InventoryClient) GetResourceTypeAlarmDictionary(resourceTypeID uuid.UUID) (*AlarmDictionary, error) {
	resp, err := client.GetResourceTypeAlarmDictionaryWithResponse(context.TODO(), resourceTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource type alarm dictionary: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get resource type alarm dictionary: received error from api: %w",
			apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// ListAlarmDictionaries lists all alarm dictionaries.
func (client *InventoryClient) ListAlarmDictionaries() ([]AlarmDictionary, error) {
	resp, err := client.GetAlarmDictionariesWithResponse(context.TODO(), &inventory.GetAlarmDictionariesParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list alarm dictionaries: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list alarm dictionaries: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// GetAlarmDictionary retrieves an alarm dictionary by its alarmDictionaryId.
func (client *InventoryClient) GetAlarmDictionary(alarmDictionaryID uuid.UUID) (*AlarmDictionary, error) {
	resp, err := client.GetAlarmDictionaryWithResponse(context.TODO(), alarmDictionaryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alarm dictionary: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get alarm dictionary: received error from api: %w", apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// ListDeploymentManagers lists all deployment managers. Optionally, list parameters can be provided.
func (client *InventoryClient) ListDeploymentManagers(params ...*InventoryListParams) ([]DeploymentManager, error) {
	var listParams *InventoryListParams
	if len(params) > 0 {
		listParams = params[0]
	}

	resp, err := client.GetDeploymentManagersWithResponse(context.TODO(), listParams.deploymentManagersParams())
	if err != nil {
		return nil, fmt.Errorf("failed to list deployment managers: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list deployment managers: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// GetDeploymentManager retrieves a deployment manager by its deploymentManagerId.
func (client *InventoryClient) GetDeploymentManager(deploymentManagerID uuid.UUID) (*DeploymentManager, error) {
	resp, err := client.GetDeploymentManagerWithResponse(context.TODO(), deploymentManagerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment manager: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to get deployment manager: received error from api: %w", apiErrorFromResponse(resp))
	}

	return resp.JSON200, nil
}

// ListSubscriptions lists all inventory subscriptions.
func (client *InventoryClient) ListSubscriptions() ([]InventorySubscription, error) {
	klog.V(100).Info("Listing inventory subscriptions")

	resp, err := client.GetSubscriptionsWithResponse(context.TODO(), &inventory.GetSubscriptionsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return nil, fmt.Errorf("failed to list subscriptions: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// CreateSubscription creates a new inventory subscription.
func (client *InventoryClient) CreateSubscription(subscription InventorySubscription) (InventorySubscription, error) {
	klog.V(100).Infof("Creating inventory subscription %#v", subscription)

	resp, err := client.CreateSubscriptionWithResponse(context.TODO(), subscription)
	if err != nil {
		return InventorySubscription{}, fmt.Errorf("failed to create subscription: error contacting api: %w", err)
	}

	if resp.StatusCode() != 201 || resp.JSON201 == nil {
		return InventorySubscription{},
			fmt.Errorf("failed to create subscription: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON201, nil
}

// GetSubscription retrieves an inventory subscription by its subscriptionId.
func (client *InventoryClient) GetSubscription(id uuid.UUID) (InventorySubscription, error) {
	klog.V(100).Infof("Getting inventory subscription with id %v", id)

	resp, err := client.GetSubscriptionWithResponse(context.TODO(), id)
	if err != nil {
		return InventorySubscription{}, fmt.Errorf("failed to get subscription: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 || resp.JSON200 == nil {
		return InventorySubscription{},
			fmt.Errorf("failed to get subscription: received error from api: %w", apiErrorFromResponse(resp))
	}

	return *resp.JSON200, nil
}

// DeleteSubscription deletes an inventory subscription by its subscriptionId.
func (client *InventoryClient) DeleteSubscription(id uuid.UUID) error {
	klog.V(100).Infof("Deleting inventory subscription with id %v", id)

	resp, err := client.DeleteSubscriptionWithResponse(context.TODO(), id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: error contacting api: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to delete subscription: received error from api: %w", apiErrorFromResponse(resp))
	}

	return nil
}
