package objectcreate

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/require"

	// appsv1 "k8s.io/api/apps/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	// "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestObjectPlugin(t *testing.T) {
	name, namespace := "name", "namespace"
	tests := []struct {
		name     string
		objName  string
		capsule  *v1alpha2.Capsule
		current  *unstructured.Unstructured
		config   string
		expected *unstructured.Unstructured
	}{
		{
			name:    "simple object",
			capsule: &v1alpha2.Capsule{},
			current: nil,
			objName: name,
			config: `
object: |
  apiVersion: vpcresources.k8s.aws/v1beta1
  kind: SecurityGroupPolicy
  spec:
    podSelector:
      matchLabels:
        app: green-pod
`,
			expected: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "vpcresources.k8s.aws/v1beta1",
					"kind":       "SecurityGroupPolicy",
					"metadata": map[string]any{
						"name":      name,
						"namespace": namespace,
					},
					"spec": map[string]any{
						"podSelector": map[string]any{
							"matchLabels": map[string]any{
								"app": "green-pod",
							},
						},
					},
				},
			},
		},
		{
			name: "using capsule annotations",
			capsule: &v1alpha2.Capsule{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						"groupIDs": "[id1, id2]",
					},
				},
			},
			objName: "someobj",
			current: nil,
			config: `
object: |
  apiVersion: vpcresources.k8s.aws/v1beta1
  kind: SecurityGroupPolicy
  metadata:
    name: someobj
  spec:
    podSelector:
      matchLabels:
        app: green-pod
      securityGroups:
        groupIds: {{ .capsule.metadata.annotations.groupIDs }}
`,
			expected: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "vpcresources.k8s.aws/v1beta1",
					"kind":       "SecurityGroupPolicy",
					"metadata": map[string]any{
						"name":      "someobj",
						"namespace": namespace,
					},
					"spec": map[string]any{
						"podSelector": map[string]any{
							"matchLabels": map[string]any{
								"app": "green-pod",
							},
							"securityGroups": map[string]any{
								"groupIds": []any{"id1", "id2"},
							},
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
			vm := scheme.NewVersionMapperFromScheme(scheme.New())
			p := pipeline.NewCapsulePipeline(nil, scheme.New(), vm, logr.FromContextOrDiscard(context.Background()))
			req := pipeline.NewCapsuleRequest(p, tt.capsule, nil)
			if tt.current != nil {
				require.NoError(t, req.Set(tt.current))
			}

			plugin := Plugin{
				configBytes: []byte(tt.config),
			}
			require.NoError(t, plugin.Run(context.Background(), req, hclog.Default()))
			vpa := &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "vpcresources.k8s.aws/v1beta1",
					"kind":       "SecurityGroupPolicy",
					"metadata": map[string]any{
						"name":      tt.objName,
						"namespace": namespace,
					},
				},
			}
			require.NoError(t, req.GetNewInto(vpa))
			require.Equal(t, tt.expected, vpa)
		})
	}
}
