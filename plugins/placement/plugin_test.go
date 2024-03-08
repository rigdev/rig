package main

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
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
		stepName    string
		pluginName  string

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
requirePluginName: true`,
			annotations: map[string]string{
				PluginNameAnnotation: "other-id",
			},
			pluginName: "pluginName",
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
requirePluginName: true
requireStepName: true`,
			annotations: map[string]string{
				PluginNameAnnotation: "pluginName",
				StepNameAnnotation:   "stepName",
			},
			pluginName: "pluginName",
			stepName:   "stepName",
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
			p := pipeline.New(nil, nil, nil, scheme.New(), logr.FromContextOrDiscard(context.Background()))
			if tt.annotations == nil {
				tt.annotations = map[string]string{}
			}
			req := pipeline.NewCapsuleRequest(p, &v1alpha2.Capsule{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.annotations,
				},
			})
			require.NoError(t, req.Set(&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: tt.podSpec,
					},
				},
			}))
			pp := &placement{
				configBytes: []byte(tt.config),
				stepName:    tt.stepName,
				pluginName:  tt.pluginName,
			}
			assert.NoError(t, pp.Run(context.Background(), req, hclog.Default()))
			deploy := &appsv1.Deployment{}
			assert.NoError(t, req.GetNew(deploy))
			assert.Equal(t, tt.expected, deploy.Spec.Template.Spec)
		})
	}
}
