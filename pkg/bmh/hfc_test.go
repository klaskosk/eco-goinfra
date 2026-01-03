package bmh

import (
	"testing"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

func Test_PullHFC(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullHFC,
		bmhv1alpha1.AddToScheme,
		bmhv1alpha1.GroupVersion.WithKind("HostFirmwareComponents"),
	).ExecuteTests(t)
}

func Test_HFCBuilder_Methods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[bmhv1alpha1.HostFirmwareComponents, HFCBuilder](
		bmhv1alpha1.AddToScheme,
		bmhv1alpha1.GroupVersion.WithKind("HostFirmwareComponents"),
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		Run(t)
}
