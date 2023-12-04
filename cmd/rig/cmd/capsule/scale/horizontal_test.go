package scale

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseLabelSelectors(t *testing.T) {
	tests := []struct {
		name     string
		inp      string
		expected map[string]string
		err      bool
	}{
		{
			name:     "empty",
			inp:      "",
			expected: nil,
		},
		{
			name: "one value",
			inp:  "KEY=value",
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name: "multiple values",
			inp:  "  key=value  key2=value2 \tkey3=value3 \t",
			expected: map[string]string{
				"key":  "value",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "fail",
			inp:  "  key=value  key2=value2 \tkey3=value3 hej",
			err:  true,
		},
		{
			name: "fail - malformed key",
			inp:  "!12key=value",
			err:  true,
		},
		{
			name: "fail - malformed value",
			inp:  "key=!value",
			err:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseLabelSelectors(tt.inp)
			assert.Equal(t, err != nil, tt.err)
			assert.Equal(t, tt.expected, res)
		})
	}
}
