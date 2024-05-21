package placement

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Plugin(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		podSpec     corev1.PodSpec
		config      string
		tag         string

		expected corev1.PodSpec
	}{
		{
			name: "match all",
			podSpec: corev1.PodSpec{
				NodeSelector: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			config: `
nodeSelector:
  key1: value3
  key3: value4`,
			expected: corev1.PodSpec{
				NodeSelector: map[string]string{
					"key1": "value3",
					"key2": "value2",
					"key3": "value4",
				},
			},
		},
		{
			name: "dont match",
			podSpec: corev1.PodSpec{
				NodeSelector: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			config: `
nodeSelector:
  key1: value3
  key3: value4
requireTag: true`,
			annotations: map[string]string{
				corev1.PreferAvoidPodsAnnotationKey: "other-id",
			},
			tag: "pluginID",
			expected: corev1.PodSpec{
				NodeSelector: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		{
			name:    "match annotation",
			podSpec: corev1.PodSpec{},
			config: `
tolerations:
  - key: tol
    value: val
requireTag: true
requireTag: true`,
			annotations: map[string]string{
				TagAnnotation: "stepID",
			},
			tag: "stepID",
			expected: corev1.PodSpec{
				NodeSelector: map[string]string{},
				Tolerations: []corev1.Toleration{{
					Key:   "tol",
					Value: "val",
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := scheme.NewVersionMapperFromScheme(scheme.New())
			p := pipeline.NewCapsulePipeline(nil, scheme.New(), vm, logr.FromContextOrDiscard(context.Background()))
			if tt.annotations == nil {
				tt.annotations = map[string]string{}
			}
			req := pipeline.NewCapsuleRequest(p, &v1alpha2.Capsule{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			}, nil)
			require.NoError(t, req.Set(&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: tt.podSpec,
					},
				},
			}))
			pp := &Plugin{
				configBytes: []byte(tt.config),
				tag:         tt.tag,
			}
			assert.NoError(t, pp.Run(context.Background(), req, hclog.Default()))
			deploy := &appsv1.Deployment{}
			assert.NoError(t, req.GetNewInto(deploy))
			assert.Equal(t, tt.expected, deploy.Spec.Template.Spec)
		})
	}
}
