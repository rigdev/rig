package controller

import (
	"context"
	"fmt"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

type VPAStep struct {
	cfg *v1alpha1.OperatorConfig
}

func NewVPAStep(cfg *v1alpha1.OperatorConfig) *VPAStep {
	return &VPAStep{
		cfg: cfg,
	}
}

func (s *VPAStep) Apply(_ context.Context, req pipeline.CapsuleRequest) error {
	if !s.cfg.VerticalPodAutoscaler.Enabled {
		return nil
	}

	vpa := s.createVPA(req)
	return req.Set(vpa)
}

func (s *VPAStep) createVPA(req pipeline.CapsuleRequest) *vpav1.VerticalPodAutoscaler {
	vpa := &vpav1.VerticalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind: "VerticalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
		},
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind: "Deployment",
				Name: req.Capsule().Name,
			},
			UpdatePolicy: &vpav1.PodUpdatePolicy{
				UpdateMode: ptr.New(vpav1.UpdateModeOff),
			},
			ResourcePolicy: &vpav1.PodResourcePolicy{
				ContainerPolicies: []vpav1.ContainerResourcePolicy{{
					ControlledResources: &[]v1.ResourceName{v1.ResourceMemory},
				}},
			},
			Recommenders: []*vpav1.VerticalPodAutoscalerRecommenderSelector{{
				Name: "default", // Use a specific name once we create our own VPA recommenders
			}},
		},
	}

	return vpa
}

// This should be used once we create a VPA per namespace
func (s *VPAStep) createVPARecommender(req pipeline.CapsuleRequest) *appsv1.Deployment { //nolint:unused
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-vpa", req.Capsule().Namespace),
			Namespace: "kube-system",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": req.Capsule().Namespace,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": req.Capsule().Namespace,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "vpa-recommender",
					Containers: []v1.Container{{
						Name:    "recommender",
						Image:   "registry.k8s.io/autoscaling/vpa-recommender:1.0.0",
						Command: []string{"/recommender"},
						Args: []string{
							"--recommender-name", req.Capsule().Namespace,
							"--vpa-object-namespace", req.Capsule().Namespace,
						},
						Ports: []v1.ContainerPort{{
							Name:          "prometheus",
							ContainerPort: 8492,
						}},
						Resources: v1.ResourceRequirements{
							Requests: map[v1.ResourceName]resource.Quantity{
								v1.ResourceCPU:    resource.MustParse("50m"),
								v1.ResourceMemory: resource.MustParse("500Mi"),
							},
						},
						SecurityContext: &v1.SecurityContext{
							RunAsUser:    ptr.New(int64(65534)),
							RunAsNonRoot: ptr.New(true),
						},
					}},
				},
			},
		},
	}
}
