package utils

import (
	"testing"

	"github.com/rigdev/rig/pkg/errors"
)

func TestValiateConfigFilePath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected error
	}{
		{
			name:     "empty path",
			path:     "",
			expected: errors.InvalidArgumentErrorf("must not be empty"),
		},
		{
			name:     "relative path",
			path:     "config/config.yaml",
			expected: errors.InvalidArgumentErrorf("must be an absolute path"),
		},
		{
			name:     "path with whitespace",
			path:     "/path/with whitespace",
			expected: errors.InvalidArgumentErrorf("must not contain unescaped whitespace"),
		},
		{
			name:     "path with escaped whitespace",
			path:     "/path/with\\ whitespace",
			expected: nil,
		},
		{
			name:     "path with double slashes",
			path:     "/path//with/double/slashes",
			expected: errors.InvalidArgumentErrorf("must not contain double slashes"),
		},
		{
			name:     "path with dots",
			path:     "/path/with/../dots",
			expected: errors.InvalidArgumentErrorf("must not contain dots"),
		},
		{
			name:     "valid path",
			path:     "/path/to/config.yaml",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValiateConfigFilePath(tc.path)
			if errors.CodeOf(err) != errors.CodeOf(tc.expected) && errors.MessageOf(err) != errors.MessageOf(tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, err)
			}
		})
	}
}
