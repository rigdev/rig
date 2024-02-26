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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Plugin(t *testing.T) {
	name, namespace := "name", "namespace"
	tests := []struct {
		name              string
		capsule           *v1alpha2.Capsule
		annotations       map[string]string
		annotationsBefore map[string]string
		expected          map[string]string
	}{
		{
			name: "add delete remove annotation",
			capsule: &v1alpha2.Capsule{
				Spec: v1alpha2.CapsuleSpec{},
			},
			annotations: map[string]string{
				"annotation1": "",
				"annotation2": "new-value",
				"annotation4": "value4",
			},
			annotationsBefore: map[string]string{
				"annotation1": "value1",
				"annotation2": "value2",
				"annotation3": "value3",
			},
			expected: map[string]string{
				"annotation2": "new-value",
				"annotation3": "value3",
				"annotation4": "value4",
			},
		},
		{
			name: "capsule templating",
			capsule: &v1alpha2.Capsule{
				Spec: v1alpha2.CapsuleSpec{Image: "nginx"},
			},
			annotations: map[string]string{
				"annotation": "image-{{ .capsule.spec.image }}",
			},
			annotationsBefore: map[string]string{},
			expected: map[string]string{
				"annotation": "image-nginx",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.capsule.Namespace = namespace
			tt.capsule.Name = name
			p := pipeline.New(nil, nil, scheme.New(), logr.FromContextOrDiscard(context.Background()))
			req := pipeline.NewCapsuleRequest(p, tt.capsule)
			req.Set(&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Annotations: tt.annotationsBefore,
				},
			})
			plugin := annotationsPlugin{
				config: Config{
					Annotations: tt.annotations,
					Group:       "apps",
					Kind:        "Deployment",
					Name:        name,
				},
			}
			err := plugin.Run(context.Background(), req, hclog.Default())
			assert.Nil(t, err)
			deploy := &appsv1.Deployment{}
			err = req.GetNew(deploy)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, deploy.GetAnnotations())
		})
	}
}