package k8s

import (
	"strconv"
	"testing"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/stretchr/testify/assert"
)

func TestCreateProxyPorts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in       []*capsule.Interface
		expected []uint32
	}{
		{
			in:       []*capsule.Interface{},
			expected: []uint32{},
		},
		{
			in:       []*capsule.Interface{{}, {}, {}},
			expected: []uint32{49152, 49153, 49154},
		},
		{
			in:       []*capsule.Interface{{Port: 49152}, {}, {}},
			expected: []uint32{49153, 49154, 49155},
		},
		{
			in:       []*capsule.Interface{{Port: 49153}, {Port: 49152}, {}},
			expected: []uint32{49154, 49155, 49156},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			actual, err := createProxyPorts(test.in)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
