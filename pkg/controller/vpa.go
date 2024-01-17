package controller

import (
	"context"
	"fmt"

	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func reconcileVPA(ctx context.Context, r *reconcileRequest) error {
	if err := r.reconcileVPA(ctx); err != nil {
		return err
	}

	return nil
}

func (r *reconcileRequest) reconcileVPA(ctx context.Context) error {
	if !r.config.VerticalPodAutoscaler.Enabled {
		return nil
	}

	vpa, err := r.createVPA()
	if err != nil {
		return err
	}
	existingVPA := &vpav1.VerticalPodAutoscaler{}
	if err := r.client.Get(ctx, client.ObjectKeyFromObject(vpa), existingVPA); err != nil {
		if kerrors.IsNotFound(err) {
			r.logger.Info("creating vertical pod autoscaler")
			if err := r.client.Create(ctx, vpa); err != nil {
				return fmt.Errorf("could not create vertical pod autoscaler: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("could not fetch vertical pod autoscaler: %w", err)
		}
	}
	return upsertIfNewer(
		ctx,
		r,
		existingVPA,
		vpa,
		func(t1, t2 *vpav1.VerticalPodAutoscaler) bool {
			return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
		},
	)
}

func (r *reconcileRequest) createVPA() (*vpav1.VerticalPodAutoscaler, error) {
	vpa := &vpav1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.capsule.Name,
			Namespace: r.capsule.Namespace,
		},
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscalingv1.CrossVersionObjectReference{
				Kind: "Deployment",
				Name: r.capsule.Name,
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

	if err := controllerutil.SetControllerReference(&r.capsule, vpa, r.scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on vertical pod autoscaler: %w", err)
	}

	return vpa, nil
}

// This should be used once we create a VPA per namespace
func (r *reconcileRequest) reconcileVPARecommender(ctx context.Context) error { //nolint:unused
	recommender := r.createVPARecommender()
	existingRecommender := &appsv1.Deployment{}
	if err := r.client.Get(ctx, client.ObjectKeyFromObject(recommender), existingRecommender); err != nil {
		if kerrors.IsNotFound(err) {
			r.logger.Info("creating vertical pod autoscaler recommender")
			if err := r.client.Create(ctx, recommender); err != nil {
				return fmt.Errorf("could not create vertical pod autoscaler recommender: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("could not fetch vertical pod autoscaler recommender: %w", err)
		}
	}
	return upsertIfNewer(ctx, r,
		existingRecommender, recommender,
		func(t1, t2 *appsv1.Deployment) bool {
			return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
		},
	)
}

func (r *reconcileRequest) createVPARecommender() *appsv1.Deployment { //nolint:unused
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-vpa", r.capsule.Namespace),
			Namespace: "kube-system",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": r.req.Namespace,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": r.req.Namespace,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "vpa-recommender",
					Containers: []v1.Container{{
						Name:    "recommender",
						Image:   "registry.k8s.io/autoscaling/vpa-recommender:1.0.0",
						Command: []string{"/recommender"},
						Args: []string{
							"--recommender-name", r.req.Namespace,
							"--vpa-object-namespace", r.req.Namespace,
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
