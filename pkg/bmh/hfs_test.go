package bmh

import (
	"testing"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

func Test_HFSBuilder_PullHFS(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullHFS,
		bmhv1alpha1.AddToScheme,
		bmhv1alpha1.GroupVersion.WithKind("HostFirmwareSettings"),
	).ExecuteTests(t)
}

func Test_HFSBuilder_Methods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[bmhv1alpha1.HostFirmwareSettings, HFSBuilder](
		bmhv1alpha1.AddToScheme,
		bmhv1alpha1.GroupVersion.WithKind("HostFirmwareSettings"),
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}
