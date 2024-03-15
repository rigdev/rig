package googlesqlproxy

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Plugin(t *testing.T) {
	name, namespace := "name", "namespace"

	tests := []struct {
		name        string
		config      string
		deployment  *appsv1.Deployment
		expected    *appsv1.Deployment
		expectedErr error
	}{
		{
			name:   "container already exists",
			config: "{}",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "google-cloud-sql-proxy"}}},
					},
				},
			},
			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "google-cloud-sql-proxy"}}},
					},
				},
			},
			expectedErr: errors.New("there was already a container named 'google-cloud-sql-proxy'"),
		},
		{
			name: "complex",
			config: `
tag: v1
args: ['arg1', 'arg2', 'arg-{{ .capsule.metadata.name }}']
instanceConnectionNames: ['ins1', 'ins2']
resources:
  cpu: 0.1
  memory: 50M
envFromSource:
  - name: source1
    kind: ConfigMap
  - name: source2
    kind: Secret
envVars:
  - name: var1
    value: value1
files:
  - path: /some/path/file1.txt
    ref:
      kind: ConfigMap
      name: file1
      key: data
  - path: /some/path/file2.txt
    ref:
      kind: ConfigMap
      name: file2
      key: data
`,
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Volumes: []corev1.Volume{
								{
									Name: "configmap-file1",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{Name: "file1"},
											Items: []corev1.KeyToPath{{
												Key:  "data",
												Path: "file1.txt",
											}},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{{
								Name:  "google-cloud-sql-proxy",
								Image: "gcr.io/cloud-sql-connectors/cloud-sql-proxy:v1",
								Args:  []string{"ins1", "ins2", "arg1", "arg2", "arg-name"},
								EnvFrom: []corev1.EnvFromSource{
									{ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: "source1"},
									}},
									{SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{Name: "source2"},
									}},
								},
								Env: []corev1.EnvVar{{
									Name:  "var1",
									Value: "value1",
								}},
								Resources: corev1.ResourceRequirements{
									Requests: map[corev1.ResourceName]resource.Quantity{
										corev1.ResourceCPU:    resource.MustParse("0.1"),
										corev1.ResourceMemory: resource.MustParse("50M"),
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "configmap-file1",
										MountPath: "/some/path/file1.txt",
										SubPath:   "file1.txt",
									},
									{
										Name:      "configmap-file2",
										MountPath: "/some/path/file2.txt",
										SubPath:   "file2.txt",
									},
								},
								SecurityContext: &corev1.SecurityContext{
									RunAsNonRoot: ptr.New(true),
								},
								RestartPolicy: ptr.New(corev1.ContainerRestartPolicyAlways),
							}},
							Volumes: []corev1.Volume{
								{
									Name: "configmap-file1",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{Name: "file1"},
											Items: []corev1.KeyToPath{{
												Key:  "data",
												Path: "file1.txt",
											}},
										},
									},
								},
								{
									Name: "configmap-file2",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{Name: "file2"},
											Items: []corev1.KeyToPath{{
												Key:  "data",
												Path: "file2.txt",
											}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pipeline.New(nil, nil, nil, scheme.New(), logr.FromContextOrDiscard(context.Background()))
			req := pipeline.NewCapsuleRequest(p, &v1alpha2.Capsule{ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			}})
			tt.deployment.ObjectMeta.Name = name
			assert.NoError(t, req.Set(tt.deployment))
			plugin := Plugin{
				configBytes: []byte(tt.config),
			}
			err := plugin.Run(context.Background(), req, hclog.Default())
			utils.ErrorEqual(t, tt.expectedErr, err)
			deploy := &appsv1.Deployment{}
			assert.NoError(t, req.GetNew(deploy))
			assert.Equal(t, tt.expected.Spec, deploy.Spec)
		})
	}
}
