package utils

import (
	"testing"

	"github.com/rigdev/rig/pkg/errors"
	"github.com/stretchr/testify/assert"
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
		{
			name:     "empty path",
			path:     "/",
			expected: errors.InvalidArgumentErrorf("must not end with a slash"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValiateConfigFilePath(tc.path)
			if errors.CodeOf(err) != errors.CodeOf(tc.expected) || errors.MessageOf(err) != errors.MessageOf(tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, err)
			}
		})
	}
}

func Test_ValidateURLPath(t *testing.T) {
	tests := []struct {
		name string
		p    string
		err  bool
	}{
		{
			name: "empty",
			p:    "",
			err:  false,
		},
		{
			name: "one segment",
			p:    "/hej",
			err:  false,
		},
		{
			name: "multiple segments",
			p:    "/h:~ej/12path/SomEWhe_r-e@",
			err:  false,
		},
		{
			name: "bad segment",
			p:    "//hej",
			err:  true,
		},
		{
			name: "bad character",
			p:    "/hej/hej?",
			err:  true,
		},
		{
			name: "path with escape characters",
			p:    "/hej/hej%f35%0A",
			err:  false,
		},
		{
			name: "path with malformed escape characters",
			p:    "/hej/hej%f35%0",
			err:  true,
		},
		{
			name: "just a slash",
			p:    "/",
			err:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURLPath(tt.p)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
