package configmap

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

var configMapGVK = corev1.SchemeGroupVersion.WithKind("ConfigMap")

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig[corev1.ConfigMap, Builder](NewBuilder, corev1.AddToScheme, configMapGVK).
		ExecuteTests(t)
}

func TestPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig[corev1.ConfigMap, Builder](Pull, corev1.AddToScheme, configMapGVK).
		ExecuteTests(t)
}

func TestBuilderMethods(t *testing.T) {
	t.Parallel()

	commonConfig := newConfigMapCommonTestConfig()

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonConfig)).
		With(testhelper.NewExistsTestConfig(commonConfig)).
		With(testhelper.NewCreateTestConfig(commonConfig)).
		With(testhelper.NewDeleterTestConfig(commonConfig)).
		With(testhelper.NewUpdateTestConfig(commonConfig)).
		Run(t)
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	testhelper.NewWithOptionsTestConfig(newConfigMapCommonTestConfig()).ExecuteTests(t)
}

func TestWithData(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builder       func() *Builder
		data          map[string]string
		assertError   func(error) bool
		expectedData  map[string]string
		expectDataNil bool
	}{
		{
			name: "sets data on valid builder",
			builder: func() *Builder {
				return NewBuilder(newConfigMapTestClient(), "test-name", "test-namespace")
			},
			data:         map[string]string{"key": "value"},
			assertError:  func(err error) bool { return err == nil },
			expectedData: map[string]string{"key": "value"},
		},
		{
			name: "empty data sets builder error",
			builder: func() *Builder {
				return NewBuilder(newConfigMapTestClient(), "test-name", "test-namespace")
			},
			data:          map[string]string{},
			assertError:   func(err error) bool { return err != nil && err.Error() == "'data' cannot be empty" },
			expectDataNil: true,
		},
		{
			name: "invalid builder short circuits",
			builder: func() *Builder {
				return NewBuilder(newConfigMapTestClient(), "", "test-namespace")
			},
			data:          map[string]string{"key": "value"},
			assertError:   commonerrors.IsBuilderNameEmpty,
			expectDataNil: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := testCase.builder()
			require.NotNil(t, builder)

			result := builder.WithData(testCase.data)
			require.Same(t, builder, result)
			require.Truef(t, testCase.assertError(result.GetError()), "unexpected error: %v", result.GetError())

			if testCase.expectDataNil {
				assert.Nil(t, result.Definition.Data)
			} else {
				assert.Equal(t, testCase.expectedData, result.Definition.Data)
			}
		})
	}
}

func TestGetGVR(t *testing.T) {
	t.Parallel()

	testGVR := GetGVR()
	assert.Equal(t, "configmaps", testGVR.Resource)
	assert.Equal(t, "v1", testGVR.Version)
	assert.Equal(t, "", testGVR.Group)
}

// newConfigMapCommonTestConfig returns the shared testhelper configuration for configmap builder tests.
func newConfigMapCommonTestConfig() testhelper.CommonTestConfig[corev1.ConfigMap, Builder, *corev1.ConfigMap, *Builder] {
	return testhelper.NewCommonTestConfig[corev1.ConfigMap, Builder](
		corev1.AddToScheme, configMapGVK, testhelper.ResourceScopeNamespaced)
}

// newConfigMapTestClient returns a fake client configured with the ConfigMap scheme.
func newConfigMapTestClient() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: []clients.SchemeAttacher{corev1.AddToScheme},
	})
}
