/*
Copyright 2023 Rig.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	rigdevv1alpha1 "github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CapsuleReconciler reconciles a Capsule object
type CapsuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	labelRigDevCapsule = "rig.dev/capsule"
	finalizer          = "rig.dev/finalizer"
)

//+kubebuilder:rbac:groups=rig.dev,resources=capsules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rig.dev,resources=capsules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rig.dev,resources=capsules/finalizers,verbs=update

// Reconcile compares the state specified by the Capsule object against the
// actual cluster state, and then performs operations to make the cluster state
// reflect the state specified by the Capsule.
func (r *CapsuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO: use rig logger
	log := log.FromContext(ctx)

	capsule := &rigdevv1alpha1.Capsule{}
	if err := r.Get(ctx, req.NamespacedName, capsule); err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("could not fetch Capsule: %w", err)
	}

	log.Info("reconcile, hpa", capsule.Spec.HorizontalScale)

	if result, err := r.reconcileDeployment(ctx, req, log, capsule); err != nil {
		return result, err
	}
	if result, err := r.reconcileService(ctx, req, log, capsule); err != nil {
		return result, err
	}
	if result, err := r.reconcileIngress(ctx, req, log, capsule); err != nil {
		return result, err
	}
	if result, err := r.reconcileLoadBalancer(ctx, req, log, capsule); err != nil {
		return result, err
	}
	if result, err := r.reconcileLoadBalancer(ctx, req, log, capsule); err != nil {
		return result, err
	}
	if result, err := r.reconcileHorizontalPodAutoscaler(ctx, req, log, capsule); err != nil {
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *CapsuleReconciler) reconcileDeployment(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
) (ctrl.Result, error) {
	deploy, err := createDeployment(capsule, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	existingDeploy := &appsv1.Deployment{}
	if err = r.Get(ctx, req.NamespacedName, existingDeploy); err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("creating deployment")
			if err := r.Create(ctx, deploy); err != nil {
				return ctrl.Result{}, fmt.Errorf("could not create deployment: %w", err)
			}
			existingDeploy = deploy
		} else {
			return ctrl.Result{}, fmt.Errorf("could not fetch deployment: %w", err)
		}
	}

	if !IsOwnedBy(capsule, existingDeploy) {
		log.Info("Found existing deployment not owned by capsule. Will not update it.")
		return ctrl.Result{}, errors.New("found existing deployment not owned by capsule")
	}

	if !reflect.DeepEqual(existingDeploy.Spec, deploy.Spec) {
		log.Info("updating deployment")
		if err := r.Update(ctx, deploy); err != nil {
			return ctrl.Result{}, fmt.Errorf("could not update deployment: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func createDeployment(
	capsule *rigdevv1alpha1.Capsule,
	scheme *runtime.Scheme,
) (*appsv1.Deployment, error) {
	var ports []v1.ContainerPort
	for _, i := range capsule.Spec.Interfaces {
		ports = append(ports, v1.ContainerPort{
			Name:          i.Name,
			ContainerPort: i.Port,
		})
	}

	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount
	for _, f := range capsule.Spec.Files {
		var name string
		switch {
		case f.ConfigMap != nil:
			name = "volume-" + strings.ReplaceAll(f.ConfigMap.Name, ".", "-")
			volumes = append(volumes, v1.Volume{
				Name: name,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: f.ConfigMap.Name,
						},
						Items: []v1.KeyToPath{
							{
								Key:  f.ConfigMap.Key,
								Path: path.Base(f.Path),
							},
						},
					},
				},
			})
		case f.Secret != nil:
			name = "volume-" + strings.ReplaceAll(f.Secret.Name, ".", "-")
			volumes = append(volumes, v1.Volume{
				Name: name,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: f.Secret.Name,
						Items: []v1.KeyToPath{
							{
								Key:  f.Secret.Key,
								Path: path.Base(f.Path),
							},
						},
					},
				},
			})
		}
		if name != "" {
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      name,
				MountPath: f.Path,
				SubPath:   path.Base(f.Path),
			})
		}
	}

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      capsule.Name,
			Namespace: capsule.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					labelRigDevCapsule: capsule.Name,
				},
			},
			Replicas: capsule.Spec.Replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: capsule.Annotations,
					Labels: map[string]string{
						labelRigDevCapsule: capsule.Name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  capsule.Name,
							Image: capsule.Spec.Image,
							EnvFrom: []v1.EnvFromSource{
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: capsule.Name,
										},
										Optional: ptr.New(true),
									},
								},
								{
									ConfigMapRef: &v1.ConfigMapEnvSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: capsule.Name,
										},
										Optional: ptr.New(true),
									},
								},
							},
							VolumeMounts: volumeMounts,
							Ports:        ports,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(capsule, d, scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on deployment: %w", err)
	}

	return d, nil
}

func (r *CapsuleReconciler) reconcileService(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
) (ctrl.Result, error) {
	service, err := createService(capsule, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	existingService := &v1.Service{}
	if err := r.Get(ctx, req.NamespacedName, existingService); err != nil {
		if kerrors.IsNotFound(err) {
			if len(capsule.Spec.Interfaces) == 0 {
				return ctrl.Result{}, nil
			}

			log.Info("creating service")
			if err := r.Create(ctx, service); err != nil {
				return ctrl.Result{}, fmt.Errorf("could not create service: %w", err)
			}
			existingService = service
		} else {
			return ctrl.Result{}, fmt.Errorf("could not fetch service: %w", err)
		}
	}

	if !IsOwnedBy(capsule, existingService) {
		if len(capsule.Spec.Interfaces) == 0 {
			log.Info("Found existing service not owned by capsule. Will not delete it.")
		} else {
			log.Info("Found existing service not owned by capsule. Will not update it.")
			return ctrl.Result{}, errors.New("found existing service not owned by capsule")
		}
	} else {
		if len(capsule.Spec.Interfaces) == 0 {
			log.Info("deleting service")
			if err := r.Delete(ctx, existingService); err != nil {
				return ctrl.Result{}, fmt.Errorf("could not delete service: %w", err)
			}
		} else {
			if !reflect.DeepEqual(existingService.Spec, service.Spec) {
				log.Info("updating service")
				if err := r.Update(ctx, service); err != nil {
					return ctrl.Result{}, fmt.Errorf("could not update service: %w", err)
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func createService(
	capsule *rigdevv1alpha1.Capsule,
	scheme *runtime.Scheme,
) (*v1.Service, error) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      capsule.Name,
			Namespace: capsule.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				labelRigDevCapsule: capsule.Name,
			},
		},
	}

	for _, inf := range capsule.Spec.Interfaces {
		svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
			Name:       inf.Name,
			Port:       inf.Port,
			TargetPort: intstr.FromString(inf.Name),
		})
	}

	if err := controllerutil.SetControllerReference(capsule, svc, scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on service: %w", err)
	}

	return svc, nil
}

func (r *CapsuleReconciler) reconcileIngress(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
) (ctrl.Result, error) {
	ing, err := createIngress(capsule, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	existingIng := &netv1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, existingIng); err != nil {
		if kerrors.IsNotFound(err) {
			if !capsuleHasIngress(capsule) {
				return ctrl.Result{}, nil
			}

			log.Info("creating ingress")
			if err := r.Create(ctx, ing); err != nil {
				return ctrl.Result{}, fmt.Errorf("could not create ingress: %w", err)
			}
			existingIng = ing
		} else {
			return ctrl.Result{}, fmt.Errorf("could not fetch ingress: %w", err)
		}
	}

	if !IsOwnedBy(capsule, existingIng) {
		if capsuleHasIngress(capsule) {
			log.Info("Found existing ingress not owned by capsule. Will not update it.")
			return ctrl.Result{}, errors.New("found existing ingress not owned by capsule")
		} else {
			log.Info("Found existing ingress not owned by capsule. Will not delete it.")
		}
	} else {
		if capsuleHasIngress(capsule) {
			if !reflect.DeepEqual(existingIng.Spec, ing.Spec) {
				log.Info("updating ingress")
				if err := r.Update(ctx, ing); err != nil {
					return ctrl.Result{}, fmt.Errorf("could not update ingress: %w", err)
				}
			}
		} else {
			log.Info("deleting ingress")
			if err := r.Delete(ctx, existingIng); err != nil {
				return ctrl.Result{}, fmt.Errorf("could not delete ingress: %w", err)
			}
		}
	}

	return ctrl.Result{}, nil
}

func capsuleHasIngress(capsule *rigdevv1alpha1.Capsule) bool {
	for _, inf := range capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			return true
		}
	}
	return false
}

func createIngress(
	capsule *rigdevv1alpha1.Capsule,
	scheme *runtime.Scheme,
) (*netv1.Ingress, error) {
	ing := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      capsule.Name,
			Namespace: capsule.Namespace,
		},
	}

	for _, inf := range capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			// TODO: setup TLS
			ing.Spec.Rules = append(ing.Spec.Rules, netv1.IngressRule{
				Host: inf.Public.Ingress.Host,
				IngressRuleValue: netv1.IngressRuleValue{
					HTTP: &netv1.HTTPIngressRuleValue{
						Paths: []netv1.HTTPIngressPath{
							{
								PathType: ptr.New(netv1.PathTypePrefix),
								Path:     "/",
								Backend: netv1.IngressBackend{
									Service: &netv1.IngressServiceBackend{
										Name: capsule.Name,
										Port: netv1.ServiceBackendPort{
											Name: inf.Name,
										},
									},
								},
							},
						},
					},
				},
			})
		}
	}

	if err := controllerutil.SetControllerReference(capsule, ing, scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on ingress: %w", err)
	}

	return ing, nil
}

func (r *CapsuleReconciler) reconcileLoadBalancer(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
) (ctrl.Result, error) {
	svc, err := createLoadBalancer(capsule, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	nsName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-lb", req.NamespacedName.Name),
		Namespace: req.NamespacedName.Namespace,
	}
	existingSvc := &v1.Service{}
	if err := r.Get(ctx, nsName, existingSvc); err != nil {
		if kerrors.IsNotFound(err) {
			if !capsuleHasLoadBalancer(capsule) {
				return ctrl.Result{}, nil
			}

			log.Info("creating loadbalancer service")
			if err := r.Create(ctx, svc); err != nil {
				return ctrl.Result{}, fmt.Errorf("could not create loadbalancer: %w", err)
			}
			existingSvc = svc
		} else {
			return ctrl.Result{}, fmt.Errorf("could not fetch loadbalancer: %w", err)
		}
	}

	if !IsOwnedBy(capsule, existingSvc) {
		if capsuleHasLoadBalancer(capsule) {
			log.Info("Found existing loadbalancer service not owned by capsule. Will not update it.")
			return ctrl.Result{}, errors.New("found existing loadbalancer service not owned by capsule")
		} else {
			log.Info("Found existing loadbalancer service not owned by capsule. Will not delete it.")
		}
	} else {
		if capsuleHasLoadBalancer(capsule) {
			if !reflect.DeepEqual(existingSvc.Spec, svc.Spec) {
				log.Info("updating loadbalancer service")
				if err := r.Update(ctx, svc); err != nil {
					return ctrl.Result{}, fmt.Errorf("could not update loadbalancer service: %w", err)
				}
			}
		} else {
			log.Info("deleting loadbalancer service")
			if err := r.Delete(ctx, existingSvc); err != nil {
				return ctrl.Result{}, fmt.Errorf("could not delete loadbalancer service: %w", err)
			}
		}
	}

	return ctrl.Result{}, nil
}

func capsuleHasLoadBalancer(capsule *rigdevv1alpha1.Capsule) bool {
	for _, inf := range capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.LoadBalancer != nil {
			return true
		}
	}
	return false
}

func createLoadBalancer(
	capsule *rigdevv1alpha1.Capsule,
	scheme *runtime.Scheme,
) (*v1.Service, error) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-lb", capsule.Name),
			Namespace: capsule.Namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				labelRigDevCapsule: capsule.Name,
			},
		},
	}

	for _, inf := range capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.LoadBalancer != nil {
			svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
				Name:       inf.Name,
				Port:       inf.Public.LoadBalancer.Port,
				TargetPort: intstr.FromString(inf.Name),
			})
		}
	}

	if err := controllerutil.SetControllerReference(capsule, svc, scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on ingress: %w", err)
	}

	return svc, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CapsuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rigdevv1alpha1.Capsule{}).
		Complete(r)
}

func (r *CapsuleReconciler) reconcileHorizontalPodAutoscaler(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
) (ctrl.Result, error) {
	hpa, err := createHPA(capsule, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	existingHPA := &autoscalingv2.HorizontalPodAutoscaler{}
	if err = r.Get(ctx, client.ObjectKeyFromObject(hpa), existingHPA); err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("creating horizontal pod autoscaler")
			if err := r.Create(ctx, hpa); err != nil {
				return ctrl.Result{}, fmt.Errorf("could not create horizontal pod autoscaler: %w", err)
			}
			existingHPA = hpa
		} else {
			return ctrl.Result{}, fmt.Errorf("could not fetch horizontal pod autoscaler: %w", err)
		}
	}

	if !IsOwnedBy(capsule, existingHPA) {
		log.Info("Found existing horizontal pod autoscaler not owned by capsule. Will not update it.")
		return ctrl.Result{}, errors.New("found existing horizontal pod autoscaler not owned by capsule")
	}

	if !reflect.DeepEqual(existingHPA.Spec, hpa.Spec) {
		log.Info("updating hpa")
		if err := r.Update(ctx, hpa); err != nil {
			return ctrl.Result{}, fmt.Errorf("could not update deployment: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func createHPA(capsule *rigdevv1alpha1.Capsule, scheme *runtime.Scheme) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	scale := capsule.Spec.HorizontalScale

	var metrics []autoscalingv2.MetricSpec
	if scale.CPUTarget != (rigdevv1alpha1.CPUTarget{}) {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: v1.ResourceCPU,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: ptr.New(int32(scale.CPUTarget.AverageUtilizationPercentage)),
				},
			},
		})
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-autoscaler", capsule.Name),
			Namespace: capsule.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       capsule.Kind,
				Name:       capsule.Name,
				APIVersion: capsule.APIVersion,
			},
			MinReplicas: ptr.New(int32(scale.MinReplicas)),
			MaxReplicas: int32(scale.MaxReplicas),
			Metrics:     metrics,
		},
	}
	if err := controllerutil.SetControllerReference(capsule, hpa, scheme); err != nil {
		return nil, err
	}

	return hpa, nil
}
