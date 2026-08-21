package fields

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectionParams(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		selection             *Selection
		expectedFields        *string
		expectedExcludeFields *string
		expectedAllFields     *string
	}{
		{
			name:           "include",
			selection:      Include("name", "nodeClusterId"),
			expectedFields: new("name,nodeClusterId"),
		},
		{
			name:                  "exclude",
			selection:             Exclude("extensions/country"),
			expectedExcludeFields: new("extensions/country"),
		},
		{
			name:              "all",
			selection:         All(),
			expectedAllFields: new(""),
		},
		{
			name: "nil selection",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			fields, excludeFields, allFields := testCase.selection.Params()

			if testCase.expectedFields != nil {
				require.NotNil(t, fields)
				assert.Equal(t, *testCase.expectedFields, *fields)
			} else {
				assert.Nil(t, fields)
			}

			if testCase.expectedExcludeFields != nil {
				require.NotNil(t, excludeFields)
				assert.Equal(t, *testCase.expectedExcludeFields, *excludeFields)
			} else {
				assert.Nil(t, excludeFields)
			}

			if testCase.expectedAllFields != nil {
				require.NotNil(t, allFields)
				assert.Equal(t, *testCase.expectedAllFields, *allFields)
			} else {
				assert.Nil(t, allFields)
			}
		})
	}
}

func TestWithExclude(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                  string
		selection             *Selection
		exclude               []string
		expectedFields        *string
		expectedExcludeFields *string
	}{
		{
			name:                  "adds exclude to include selection",
			selection:             Include("name", "extensions"),
			exclude:               []string{"extensions/country"},
			expectedFields:        new("name,extensions"),
			expectedExcludeFields: new("extensions/country"),
		},
		{
			name:                  "nil selection delegates to exclude",
			selection:             nil,
			exclude:               []string{"extensions/country"},
			expectedExcludeFields: new("extensions/country"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := testCase.selection.WithExclude(testCase.exclude...)
			fields, excludeFields, allFields := result.Params()

			if testCase.expectedFields != nil {
				require.NotNil(t, fields)
				assert.Equal(t, *testCase.expectedFields, *fields)
			} else {
				assert.Nil(t, fields)
			}

			if testCase.expectedExcludeFields != nil {
				require.NotNil(t, excludeFields)
				assert.Equal(t, *testCase.expectedExcludeFields, *excludeFields)
			} else {
				assert.Nil(t, excludeFields)
			}

			assert.Nil(t, allFields)
		})
	}
}

func TestPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "nested field reference",
			parts:    []string{"extensions", "country"},
			expected: "extensions/country",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expected, Path(testCase.parts...))
		})
	}
}
