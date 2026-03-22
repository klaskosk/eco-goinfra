package configmap

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		nsname         string
		withClient     bool
		configMaps     []runtime.Object
		configmapCount int
		expectedError  string
	}{
		{
			name:       "returns configmaps from namespace",
			nsname:     "test-namespace",
			withClient: true,
			configMaps: []runtime.Object{
				newConfigMapObject("test-name1", "test-namespace"),
				newConfigMapObject("test-name2", "test-namespace"),
			},
			configmapCount: 2,
		},
		{
			name:       "filters configmaps from other namespaces",
			nsname:     "test-namespace",
			withClient: true,
			configMaps: []runtime.Object{
				newConfigMapObject("test-name1", "test-namespace"),
				newConfigMapObject("test-name2", "test-namespace2"),
			},
			configmapCount: 1,
		},
		{
			name:          "rejects empty namespace",
			nsname:        "",
			withClient:    true,
			expectedError: "failed to list configmaps, 'nsname' parameter is empty",
		},
		{
			name:          "rejects nil client",
			nsname:        "test-namespace",
			withClient:    false,
			expectedError: "the apiClient cannot be nil",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.withClient {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects: testCase.configMaps,
				})
			}

			configmapBuilders, err := List(testSettings, testCase.nsname)

			if testCase.expectedError != "" {
				require.Error(t, err)
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, configmapBuilders)

				return
			}

			require.NoError(t, err)
			assert.Len(t, configmapBuilders, testCase.configmapCount)

			for _, builder := range configmapBuilders {
				require.NotNil(t, builder)
				require.NotNil(t, builder.GetClient())
				require.NotNil(t, builder.GetDefinition())
				require.NotNil(t, builder.GetObject())
				assert.Equal(t, configMapGVK, builder.GetGVK())
			}
		})
	}
}

func TestListInAllNamespaces(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		withClient     bool
		configMaps     []runtime.Object
		configmapCount int
		expectedError  string
	}{
		{
			name:       "returns all configmaps",
			withClient: true,
			configMaps: []runtime.Object{
				newConfigMapObject("test-name1", "test-namespace"),
				newConfigMapObject("test-name2", "test-namespace"),
			},
			configmapCount: 2,
		},
		{
			name:       "includes multiple namespaces",
			withClient: true,
			configMaps: []runtime.Object{
				newConfigMapObject("test-name1", "test-namespace"),
				newConfigMapObject("test-name2", "test-namespace2"),
			},
			configmapCount: 2,
		},
		{
			name:          "rejects nil client",
			withClient:    false,
			expectedError: "the apiClient cannot be nil",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.withClient {
				testSettings = clients.GetTestClients(clients.TestClientParams{
					K8sMockObjects: testCase.configMaps,
				})
			}

			configmapBuilders, err := ListInAllNamespaces(testSettings)

			if testCase.expectedError != "" {
				require.Error(t, err)
				assert.EqualError(t, err, testCase.expectedError)
				assert.Nil(t, configmapBuilders)

				return
			}

			require.NoError(t, err)
			assert.Len(t, configmapBuilders, testCase.configmapCount)

			for _, builder := range configmapBuilders {
				require.NotNil(t, builder)
				require.NotNil(t, builder.GetClient())
				require.NotNil(t, builder.GetDefinition())
				require.NotNil(t, builder.GetObject())
				assert.Equal(t, configMapGVK, builder.GetGVK())
			}
		})
	}
}

// newConfigMapObject returns a ConfigMap runtime object for list tests.
func newConfigMapObject(name, namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
