package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type inp struct {
	ns      string
	capsule string
	labels  map[string]string
}

func newInp(ns, capsule string) inp {
	return inp{ns: ns, capsule: capsule}
}

func Test_Matcher(t *testing.T) {
	tests := []struct {
		name        string
		namespaces  []string
		capsules    []string
		rigplatform bool
		inputs      []inp
		expected    []bool
		selector    metav1.LabelSelector
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
		{
			name: "match labels",
			inputs: []inp{
				{
					labels: map[string]string{
						"foo": "bar",
					},
				},
				{
					labels: map[string]string{
						"foo": "baz",
					},
				},
				newInp("ns", "cap"),
			},
			expected: []bool{true, false, false},
			selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
		},
		{
			name:        "dont match rig-platform",
			rigplatform: false,
			inputs: []inp{
				newInp("ns", "cap"),
				newInp("ns", "rig-platform"),
			},
			expected: []bool{true, false},
		},
		{
			name:        "match rig-platform",
			rigplatform: true,
			inputs: []inp{
				newInp("ns", "rig-platform"),
			},
			expected: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewMatcher(tt.namespaces, tt.capsules, tt.selector, tt.rigplatform)
			assert.NoError(t, err)
			for idx, inp := range tt.inputs {
				res := matcher.Match(inp.ns, inp.capsule, inp.labels)
				assert.Equal(t, tt.expected[idx], res, "failed index %v", idx)
			}
		})
	}
}
