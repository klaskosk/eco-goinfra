package bmh

import (
	"context"
	"time"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/key"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fiveScds time.Duration = 5 * time.Second
)

// List returns bareMetalHosts inventory in the given namespace.
func List(apiClient *clients.Settings, nsname string, options ...runtimeclient.ListOptions) ([]*BmhBuilder, error) {
	if nsname == "" {
		klog.V(100).Info("bareMetalHost 'nsname' parameter can not be empty")

		return nil, commonerrors.NewBuilderFieldEmpty(
			key.NewResourceKey("BareMetalHost", "", ""), commonerrors.BuilderFieldNamespace)
	}

	convertedOptions := common.ConvertListOptionsToOptions(options)
	convertedOptions = append(convertedOptions, runtimeclient.InNamespace(nsname))

	return common.List[bmhv1alpha1.BareMetalHost, bmhv1alpha1.BareMetalHostList, BmhBuilder](
		context.TODO(), apiClient, bmhv1alpha1.AddToScheme, convertedOptions...)
}

// ListInAllNamespaces lists the BareMetalHosts across all namespaces on the provided cluster.
func ListInAllNamespaces(apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*BmhBuilder, error) {
	convertedOptions := common.ConvertListOptionsToOptions(options)

	return common.List[bmhv1alpha1.BareMetalHost, bmhv1alpha1.BareMetalHostList, BmhBuilder](
		context.TODO(), apiClient, bmhv1alpha1.AddToScheme, convertedOptions...)
}

// WaitForAllBareMetalHostsInGoodOperationalState waits for all baremetalhosts to be in good Operational State
// for a time duration up to the timeout.
func WaitForAllBareMetalHostsInGoodOperationalState(apiClient *clients.Settings,
	nsname string,
	timeout time.Duration,
	options ...runtimeclient.ListOptions) (bool, error) {
	klog.V(100).Infof("Waiting for all bareMetalHosts in %s namespace to have OK operationalStatus",
		nsname)

	bmhList, err := List(apiClient, nsname, options...)
	if err != nil {
		klog.V(100).Infof("Failed to list all bareMetalHosts in the %s namespace due to %s",
			nsname, err.Error())

		return false, err
	}

	// Wait 5 secs in each iteration before condition function () returns true or errors or times out
	// after availableDuration
	err = wait.PollUntilContextTimeout(
		context.TODO(), fiveScds, timeout, true, func(ctx context.Context) (bool, error) {
			for _, baremetalhost := range bmhList {
				status := baremetalhost.GetBmhOperationalState()

				if status != bmhv1alpha1.OperationalStatusOK {
					klog.V(100).Infof("The %s bareMetalHost in namespace %s has an unexpected operational status: %s",
						baremetalhost.Object.Name, baremetalhost.Object.Namespace, status)

					return false, nil
				}
			}

			return true, nil
		})
	if err == nil {
		klog.V(100).Infof("All baremetalhosts were found in the good Operational State "+
			"during defined timeout: %v", timeout)

		return true, nil
	}

	// Here err is "timed out waiting for the condition"
	klog.V(100).Infof("Not all baremetalhosts were found in the good Operational State "+
		"during defined timeout: %v", timeout)

	return false, err
}
