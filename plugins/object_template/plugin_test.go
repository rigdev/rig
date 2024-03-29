package objecttemplate

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestObjectPlugin(t *testing.T) {
	name, namespace := "name", "namespace"
	tests := []struct {
		name     string
		capsule  *v1alpha2.Capsule
		current  *appsv1.Deployment
		config   string
		expected *appsv1.Deployment
	}{
		{
			name:    "empty patch",
			capsule: &v1alpha2.Capsule{},
			current: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.New[int32](1),
				},
			},
			config: "object: '{}'",
			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.New[int32](1),
				},
			},
		},
		{
			name:    "overwrite replicas",
			capsule: &v1alpha2.Capsule{},
			current: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.New[int32](1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"label": "value",
						},
					},
				},
			},
			config: `
object: |
  spec:
    replicas: 2`,
			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"label": "value",
						},
					},
					Replicas: ptr.New[int32](2),
				},
			},
		},
		{
			name:    "add label",
			capsule: &v1alpha2.Capsule{},
			current: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.New[int32](1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"label": "value",
						},
					},
				},
			},
			config: `
object: |
  spec:
    selector:
      matchLabels:
        label2: value2`,
			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"label":  "value",
							"label2": "value2",
						},
					},
					Replicas: ptr.New[int32](1),
				},
			},
		},
		{
			name: "template using capsule",
			capsule: &v1alpha2.Capsule{
				Spec: v1alpha2.CapsuleSpec{
					Image: "image",
				},
			},
			current: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{},
			},
			config: `
object: |
  spec:
    selector:
      matchLabels:
        label: {{ .capsule.spec.image }}`,
			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"label": "image",
						},
					},
				},
			},
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.capsule.Namespace = namespace
			tt.capsule.Name = name
			p := pipeline.New(nil, nil, nil, scheme.New(), logr.FromContextOrDiscard(context.Background()))
			req := pipeline.NewCapsuleRequest(p, tt.capsule)
			assert.NoError(t, req.Set(tt.current))

			c := tt.config + `
group: apps
kind: Deployment
name: name`
			plugin := Plugin{
				configBytes: []byte(c),
			}
			assert.NoError(t, plugin.Run(context.Background(), req, hclog.Default()))
			deploy := &appsv1.Deployment{}
			assert.NoError(t, req.GetNew(deploy))
			tt.expected.Name = name
			tt.expected.Namespace = namespace
			assert.Equal(t, tt.expected, deploy)
		})
	}
}
