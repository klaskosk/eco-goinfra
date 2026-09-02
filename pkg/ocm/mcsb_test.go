package ocm

import (
	"testing"

	clusterv1beta2 "open-cluster-management.io/api/cluster/v1beta2"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
)

func TestNewMCSBBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewMCSBBuilder,
		clusterv1beta2.Install,
		mcsbGVK,
	).ExecuteTests(t)
}

func TestMCSBBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[clusterv1beta2.ManagedClusterSetBinding, MCSBBuilder](
		clusterv1beta2.Install,
		mcsbGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}
