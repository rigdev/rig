package obj

import (
	"testing"

	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestComparison(t *testing.T) {
	from := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "foo",
			Annotations: map[string]string{
				"something": "here",
			},
		},
	}
	to := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: "bar",
		},
	}
	c := NewComparison(from, to, scheme.New())
	c.AddFilter(RemoveAnnotationsFilter("something"))
	c.AddRemoveDiffs("metadata.name")
	d, err := c.ComputeDiff()
	require.NoError(t, err)

	require.Empty(t, d.Report.Diffs)
}
