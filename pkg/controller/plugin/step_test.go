package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type inp struct {
	ns      string
	capsule string
}

func newInp(ns, capsule string) inp {
	return inp{ns: ns, capsule: capsule}
}

func Test_Matcher(t *testing.T) {

	tests := []struct {
		name       string
		namespaces []string
		capsules   []string
		inputs     []inp
		expected   []bool
	}{
		{
			name:     "match all",
			inputs:   []inp{newInp("ns", "cap")},
			expected: []bool{true},
		},
		{
			name:       "strict match",
			namespaces: []string{"ns1", "ns2"},
			inputs: []inp{
				newInp("ns", "cap"),
				newInp("ns1", "cap"),
				newInp("ns2", "cap"),
			},
			expected: []bool{false, true, true},
		},
		{
			name:       "match prefix",
			namespaces: []string{"ns*", "ns"},
			inputs: []inp{
				newInp("ns", "cap"),
				newInp("ns2", "cap"),
				newInp("notns", "cap"),
			},
			expected: []bool{true, true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewMatcher(tt.namespaces, tt.capsules)
			assert.NoError(t, err)
			for idx, inp := range tt.inputs {
				res := matcher.Match(inp.ns, inp.capsule)
				assert.Equal(t, tt.expected[idx], res, "failed index %v", idx)
			}
		})
	}
}
