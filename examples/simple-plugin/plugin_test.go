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
	name, namespace := "capsulename", "namespace"
	tests := []struct {
		name     string
		config   string
		expected *appsv1.Deployment
	}{
		{
			name: "successfully set the label",
			config: `
label: some-label
value: some-value`,
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"some-label": "some-value",
					},
					Name:      name,
					Namespace: namespace,
				},
			},
		},
		{
			name: "set the label using the capsule name",
			config: `
label: some-label
value: {{ .capsule.metadata.name }}`,
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"some-label": "capsulename",
					},
					Name:      name,
					Namespace: namespace,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pipeline.NewCapsulePipeline(nil, nil, nil, scheme.New(), logr.FromContextOrDiscard(context.Background()))
			// Setup the CapsuleRequest
			req := pipeline.NewCapsuleRequest(p, &v1alpha2.Capsule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			})
			// Add the deployment resource to the capsule request before we execute the plugin
			// At the time plugins are executed, a Deployment resource will always be available.
			assert.NoError(t, req.Set(&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}))

			plugin := Plugin{
				configBytes: []byte(tt.config),
			}
			// Run the plugin
			assert.NoError(t, plugin.Run(context.Background(), req, hclog.Default()))

			// Extract the deployment from the CapsuleRequest
			deploy := &appsv1.Deployment{}
			assert.NoError(t, req.GetNew(deploy))

			// Check we set the expected labels
			assert.Equal(t, tt.expected, deploy)
		})
	}
}
