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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Plugin(t *testing.T) {
	name, namespace := "name", "namespace"

	tests := []struct {
		name     string
		config   Config
		expected *appsv1.Deployment
	}{
		{
			name: "set nothing",
			config: Config{
				DontAddEnabledAnnotation: true,
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    map[string]string{},
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels:      map[string]string{},
							Annotations: map[string]string{},
						},
					},
				},
			},
		},
		{
			name: "set it all",
			config: Config{
				LibraryTag: LibraryTag{
					Java:       "java",
					JavaScript: "js",
					Python:     "python",
					NET:        "net",
					Ruby:       "ruby",
				},
				UnifiedServiceTags: UnifiedServiceTags{
					Env:     "env",
					Service: "service",
					Version: "version",
				},
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
					Labels: map[string]string{
						"tags.datadoghq.com/env":     "env",
						"tags.datadoghq.com/service": "service",
						"tags.datadoghq.com/version": "version",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"admission.datadoghq.com/enabled": "true",
								"tags.datadoghq.com/env":          "env",
								"tags.datadoghq.com/service":      "service",
								"tags.datadoghq.com/version":      "version",
							},
							Annotations: map[string]string{
								"admission.datadoghq.com/java-lib.version":   "java",
								"admission.datadoghq.com/js-lib.version":     "js",
								"admission.datadoghq.com/python-lib.version": "python",
								"admission.datadoghq.com/dotnet-lib.version": "net",
								"admission.datadoghq.com/ruby-lib.version":   "ruby",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pipeline.New(nil, nil, scheme.New(), logr.FromContextOrDiscard(context.Background()))
			req := pipeline.NewCapsuleRequest(p, &v1alpha2.Capsule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			})
			assert.NoError(t, req.Set(&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}))
			plugin := datadogPlugin{
				config: tt.config,
			}
			err := plugin.Run(context.Background(), req, hclog.Default())
			assert.Nil(t, err)
			deploy := &appsv1.Deployment{}
			err = req.GetNew(deploy)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, deploy)
		})
	}
}
