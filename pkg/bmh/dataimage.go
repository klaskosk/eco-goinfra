package bmh

import (
	"context"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DataImageBuilder provides struct for the dataimage object containing connection to
// the cluster and the dataimage definitions.
type DataImageBuilder struct {
	common.EmbeddableBuilder[bmhv1alpha1.DataImage, *bmhv1alpha1.DataImage]
	common.EmbeddableDeleteReturner[bmhv1alpha1.DataImage, DataImageBuilder, *bmhv1alpha1.DataImage, *DataImageBuilder]
}

// AttachMixins attaches the mixins to the builder. This will be called automatically when the builder is initialized
// and is required for the Delete method to work.
func (b *DataImageBuilder) AttachMixins() {
	b.SetBase(b)
}

// PullDataImage retrieves an existing DataImage resource from the cluster.
func PullDataImage(apiClient *clients.Settings, name, nsname string) (*DataImageBuilder, error) {
	return common.PullNamespacedBuilder[bmhv1alpha1.DataImage, DataImageBuilder](
		context.TODO(), apiClient, bmhv1alpha1.AddToScheme, name, nsname)
}

// GetGVK returns the GVK for the DataImage resource.
func (b *DataImageBuilder) GetGVK() schema.GroupVersionKind {
	return bmhv1alpha1.GroupVersion.WithKind("DataImage")
}
