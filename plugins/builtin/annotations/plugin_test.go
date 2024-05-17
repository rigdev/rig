package annotations

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Plugin(t *testing.T) {
	name, namespace := "name", "namespace"
	tests := []struct {
		name                string
		capsule             *v1alpha2.Capsule
		config              string
		annotationsBefore   map[string]string
		expectedAnnotations map[string]string

		labelsBefore   map[string]string
		expectedLabels map[string]string
	}{
		{
			name: "add delete remove annotation",
			capsule: &v1alpha2.Capsule{
				Spec: v1alpha2.CapsuleSpec{},
			},
			config: `
annotations:
  annotation1: ""
  annotation2: "new-value"
  annotation4: "value4"`,
			annotationsBefore: map[string]string{
				"annotation1": "value1",
				"annotation2": "value2",
				"annotation3": "value3",
			},
			expectedAnnotations: map[string]string{
				"annotation2": "new-value",
				"annotation3": "value3",
				"annotation4": "value4",
			},
			expectedLabels: map[string]string{},
		},
		{
			name: "add delete remove label",
			capsule: &v1alpha2.Capsule{
				Spec: v1alpha2.CapsuleSpec{},
			},
			config: `
labels:
  label1: ""
  label2: "new-value"
  label4: "value4"`,
			labelsBefore: map[string]string{
				"label1": "value1",
				"label2": "value2",
				"label3": "value3",
			},
			expectedLabels: map[string]string{
				"label2": "new-value",
				"label3": "value3",
				"label4": "value4",
			},
			expectedAnnotations: map[string]string{},
		},
		{
			name: "capsule templating",
			capsule: &v1alpha2.Capsule{
				Spec: v1alpha2.CapsuleSpec{Image: "nginx"},
			},
			config: `
annotations:
  annotation: image-{{ .capsule.spec.image }}`,
			annotationsBefore: map[string]string{},
			expectedAnnotations: map[string]string{
				"annotation": "image-nginx",
			},
			expectedLabels: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.capsule.Namespace = namespace
			tt.capsule.Name = name
			p := pipeline.NewCapsulePipeline(nil, scheme.New(), logr.FromContextOrDiscard(context.Background()))
			req := pipeline.NewCapsuleRequest(p, tt.capsule, nil)
			assert.NoError(t, req.Set(&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Annotations: tt.annotationsBefore,
					Labels:      tt.labelsBefore,
				},
			}))
			c := tt.config + `
group: apps
kind: Deployment
name: name`
			pp := &Plugin{configBytes: []byte(c)}
			assert.NoError(t, pp.Run(context.Background(), req, hclog.Default()))
			deploy := &appsv1.Deployment{}
			assert.NoError(t, req.GetNewInto(deploy))
			assert.Equal(t, tt.expectedAnnotations, deploy.GetAnnotations())
			assert.Equal(t, tt.expectedLabels, deploy.GetLabels())
		})
	}
}
