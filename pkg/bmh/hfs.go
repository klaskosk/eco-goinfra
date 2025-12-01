package bmh

import (
	"context"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// HFSBuilder provides a struct to interface with HostFirmwareSettings resources on a specific cluster.
type HFSBuilder struct {
	common.EmbeddableBuilder[bmhv1alpha1.HostFirmwareSettings, *bmhv1alpha1.HostFirmwareSettings]
	common.EmbeddableCreator[bmhv1alpha1.HostFirmwareSettings, HFSBuilder, *bmhv1alpha1.HostFirmwareSettings, *HFSBuilder]
	common.EmbeddableDeleter[bmhv1alpha1.HostFirmwareSettings, *bmhv1alpha1.HostFirmwareSettings]
}

// AttachMixins attaches the mixins to the builder. This will be called automatically when the builder is initialized
// and is required for the Create and Delete methods to work.
func (b *HFSBuilder) AttachMixins() {
	b.EmbeddableCreator.SetBase(b)
	b.EmbeddableDeleter.SetBase(b)
}

// PullHFS pulls an existing HostFirmwareSettings from the cluster.
func PullHFS(apiClient *clients.Settings, name, nsname string) (*HFSBuilder, error) {
	return common.PullNamespacedBuilder[bmhv1alpha1.HostFirmwareSettings, HFSBuilder](
		context.TODO(), apiClient, bmhv1alpha1.AddToScheme, name, nsname)
}

// GetGVK returns the GVK for the HostFirmwareSettings resource.
func (b *HFSBuilder) GetGVK() schema.GroupVersionKind {
	return bmhv1alpha1.GroupVersion.WithKind("HostFirmwareSettings")
}
