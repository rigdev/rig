package argorollout

import (
	"context"
	"testing"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_Plugin(t *testing.T) {
	name, namespace := "name", "namespace"

	tests := []struct {
		name       string
		config     string
		deployment *appsv1.Deployment
		expected   *v1alpha1.Rollout
	}{
		{
			name:   "no config",
			config: "",
			deployment: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.New[int32](42),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"rig.dev/capsule": name,
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: namespace,
							Labels: map[string]string{
								"rig.dev/capsule": name,
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  name,
								Image: "nginx",
							}},
						},
					},
					MinReadySeconds: 1,
				},
			},
		},
		{
			name: "set strategy",
			config: `strategy:
  canary:
    tests:
      - predefined:
          name: http-success-rate
          args:
            - name: thresholdPercentage
              value: "99.99"
    steps:
      - setWeight: 10
      - pause: { duration: 30s }
      - setWeight: 20
      - pause: { duration: 30s }
      - setWeight: 100`,
			deployment: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.New[int32](42),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"rig.dev/capsule": name,
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: namespace,
							Labels: map[string]string{
								"rig.dev/capsule": name,
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  name,
								Image: "nginx",
							}},
						},
					},
					MinReadySeconds: 1,
				},
			},
			expected: &v1alpha1.Rollout{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Rollout",
					APIVersion: "argoproj.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
				Spec: v1alpha1.RolloutSpec{
					Replicas: ptr.New[int32](42),
					WorkloadRef: &v1alpha1.ObjectRef{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       name,
						ScaleDown:  "progressively",
					},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{},
						},
					},
					Strategy: v1alpha1.RolloutStrategy{
						Canary: &v1alpha1.CanaryStrategy{
							Steps: []v1alpha1.CanaryStep{
								{SetWeight: ptr.New[int32](10)},
								{Pause: &v1alpha1.RolloutPause{
									Duration: &intstr.IntOrString{
										Type:   intstr.String,
										StrVal: "30s",
									},
								}},
								{SetWeight: ptr.New[int32](20)},
								{Pause: &v1alpha1.RolloutPause{
									Duration: &intstr.IntOrString{
										Type:   intstr.String,
										StrVal: "30s",
									},
								}},
								{SetWeight: ptr.New[int32](100)},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := scheme.NewVersionMapperFromScheme(scheme.New())
			scheme := scheme.New()
			require.NoError(t, v1alpha1.AddToScheme(scheme))

			pipe := pipeline.NewCapsulePipeline(nil, scheme, vm, logr.FromContextOrDiscard(context.Background()))
			req := pipeline.NewCapsuleRequest(pipe, &v1alpha2.Capsule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			}, nil)
			require.NoError(t, req.Set(tt.deployment))
			plug := Plugin{
				configBytes: []byte(tt.config),
			}
			require.NoError(t, plug.Run(context.Background(), req, hclog.Default()))

			if tt.expected != nil {
				tt.deployment.Spec.Replicas = ptr.New[int32](0)
				rollout := &v1alpha1.Rollout{}
				require.NoError(t, req.GetNewInto(rollout))
				require.Equal(t, tt.expected, rollout)
			} else {
				require.Error(t, req.GetNewInto(&v1alpha1.Rollout{}))
			}

			deploy := &appsv1.Deployment{}
			require.NoError(t, req.GetNewInto(deploy))
			require.Equal(t, tt.deployment, deploy)
		})
	}
}
