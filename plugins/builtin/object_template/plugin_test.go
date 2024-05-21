package objecttemplate

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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
		{
			name: "change container",
			capsule: &v1alpha2.Capsule{
				Spec: v1alpha2.CapsuleSpec{
					Image: "image",
				},
			},
			current: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name: "name",
								},
							},
						},
					},
				},
			},
			config: `
object: |
  spec:
    template:
      spec:
        containers:
        - name: {{ .capsule.metadata.name }}
          image: {{ .capsule.metadata.name }}:latest`,
			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "name",
									Image: "name:latest",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "conditional match",
			capsule: &v1alpha2.Capsule{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"foo/bar": "true",
					},
				},
				Spec: v1alpha2.CapsuleSpec{
					Image: "image",
				},
			},
			current: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{},
			},
			config: `
object: |
{{ with .capsule.metadata.annotations }}
{{ if eq (index . "foo/bar") "true" }}
  spec:
    replicas: 1
{{ end }}
{{ end }}`,

			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.New(int32(1)),
				},
			},
		},
		{
			name: "conditional no match",
			capsule: &v1alpha2.Capsule{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"foo/bar": "false",
					},
				},
				Spec: v1alpha2.CapsuleSpec{
					Image: "image",
				},
			},
			current: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{},
			},
			config: `
object: |
{{ with .capsule.metadata.annotations }}
{{ if eq (index . "foo/bar") "true" }}
  spec:
    replicas: 1
{{ end }}
{{ end }}`,

			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{},
			},
		},
		{
			name: "conditional missing",
			capsule: &v1alpha2.Capsule{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: v1alpha2.CapsuleSpec{
					Image: "image",
				},
			},
			current: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{},
			},
			config: `
object: |
{{ with .capsule.metadata.annotations }}
{{ if eq (index . "foo/bar") "true" }}
  spec:
    replicas: 1
{{ end }}
{{ end }}`,

			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{},
			},
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.capsule.Namespace = namespace
			tt.capsule.Name = name
			vm := scheme.NewVersionMapperFromScheme(scheme.New())
			p := pipeline.NewCapsulePipeline(nil, scheme.New(), vm, logr.FromContextOrDiscard(context.Background()))
			req := pipeline.NewCapsuleRequest(p, tt.capsule, nil)
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
			assert.NoError(t, req.GetNewInto(deploy))
			tt.expected.Name = name
			tt.expected.Namespace = namespace
			assert.Equal(t, tt.expected, deploy)
		})
	}
}

func Test_ObjectPlugin_with_list(t *testing.T) {
	vm := scheme.NewVersionMapperFromScheme(scheme.New())
	p := pipeline.NewCapsulePipeline(nil, scheme.New(), vm, logr.FromContextOrDiscard(context.Background()))
	req := pipeline.NewCapsuleRequest(p, &v1alpha2.Capsule{}, nil)

	objects := []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "obj1"},
			Spec: appsv1.DeploymentSpec{
				Replicas:        ptr.New(int32(1)),
				MinReadySeconds: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "obj2"},
			Spec: appsv1.DeploymentSpec{
				Replicas:        ptr.New(int32(3)),
				MinReadySeconds: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "obj3"},
			Spec: appsv1.DeploymentSpec{
				Replicas:        ptr.New(int32(2)),
				MinReadySeconds: 3,
			},
		},
	}
	for _, obj := range objects {
		require.NoError(t, req.Set(obj))
	}

	config := `
group: apps
kind: Deployment
name: '*'
object: |
  spec:
    replicas: 10
    selector:
      matchLabels:
        name-{{ .current.metadata.name }}: '{{ .current.spec.minReadySeconds }}'
`
	plugin := Plugin{
		configBytes: []byte(config),
	}
	require.NoError(t, plugin.Run(context.Background(), req, hclog.Default()))

	objs, err := req.ListNew(appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind())
	require.NoError(t, err)
	deployments, err := pipeline.ListConvert[*appsv1.Deployment](objs)
	require.NoError(t, err)

	require.Equal(t, []*appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "obj1"},
			Spec: appsv1.DeploymentSpec{
				Replicas:        ptr.New(int32(10)),
				MinReadySeconds: 1,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"name-obj1": "1",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "obj2"},
			Spec: appsv1.DeploymentSpec{
				Replicas:        ptr.New(int32(10)),
				MinReadySeconds: 2,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"name-obj2": "2",
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "obj3"},
			Spec: appsv1.DeploymentSpec{
				Replicas:        ptr.New(int32(10)),
				MinReadySeconds: 3,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"name-obj3": "3",
					},
				},
			},
		},
	}, deployments)
}
