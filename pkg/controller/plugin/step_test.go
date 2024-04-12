package plugin

import (
	"testing"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/stretchr/testify/assert"
)

type inp struct {
	ns          string
	capsule     string
	annotations map[string]string
}

func newInp(ns, capsule string) inp {
	return inp{ns: ns, capsule: capsule}
}

func Test_Matcher(t *testing.T) {
	tests := []struct {
		name     string
		match    v1alpha1.CapsuleMatch
		inputs   []inp
		expected []bool
	}{
		{
			name:     "match all",
			inputs:   []inp{newInp("ns", "cap")},
			expected: []bool{true},
		},
		{
			name: "strict match",
			match: v1alpha1.CapsuleMatch{
				Namespaces: []string{"ns1", "ns2"},
			},
			inputs: []inp{
				newInp("ns", "cap"),
				newInp("ns1", "cap"),
				newInp("ns2", "cap"),
			},
			expected: []bool{false, true, true},
		},
		{
			name: "match prefix",
			match: v1alpha1.CapsuleMatch{
				Namespaces: []string{"ns*", "ns"},
			},
			inputs: []inp{
				newInp("ns", "cap"),
				newInp("ns2", "cap"),
				newInp("notns", "cap"),
			},
			expected: []bool{true, true, false},
		},
		{
			name: "match annotations",
			inputs: []inp{
				{
					annotations: map[string]string{
						"foo": "bar",
					},
				},
				{
					annotations: map[string]string{
						"foo": "baz",
					},
				},
				newInp("ns", "cap"),
			},
			expected: []bool{true, false, false},
			match: v1alpha1.CapsuleMatch{
				Annotations: map[string]string{
					"foo": "bar",
				},
			},
		},
		{
			name: "dont match rig-platform",
			match: v1alpha1.CapsuleMatch{
				EnableForPlatform: false,
			},
			inputs: []inp{
				newInp("ns", "cap"),
				newInp("ns", "rig-platform"),
			},
			expected: []bool{true, false},
		},
		{
			name: "match rig-platform",
			match: v1alpha1.CapsuleMatch{
				EnableForPlatform: true,
			},
			inputs: []inp{
				newInp("ns", "rig-platform"),
			},
			expected: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewMatcher(tt.match)
			assert.NoError(t, err)
			for idx, inp := range tt.inputs {
				res := matcher.Match(inp.ns, inp.capsule, inp.annotations)
				assert.Equal(t, tt.expected[idx], res, "failed index %v", idx)
			}
		})
	}
}
