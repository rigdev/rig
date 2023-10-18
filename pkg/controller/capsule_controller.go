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
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/go-logr/logr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	rigdevv1alpha1 "github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
	Config *configv1alpha1.OperatorConfig

	reconcileSteps []reconcileStepFunc
}

type reconcileStepFunc func(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
	status *rigdevv1alpha1.CapsuleStatus,
) error

const (
	labelRigDevCapsule = "rig.dev/capsule"
	finalizer          = "rig.dev/finalizer"
)

// SetupWithManager sets up the controller with the Manager.
func (r *CapsuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.reconcileSteps = []reconcileStepFunc{
		r.reconcileHorizontalPodAutoscaler,
		r.reconcileDeployment,
		r.reconcileService,
		r.reconcileCertificate,
		r.reconcileIngress,
		r.reconcileLoadBalancer,
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&rigdevv1alpha1.Capsule{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=rig.dev,resources=capsules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rig.dev,resources=capsules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rig.dev,resources=capsules/finalizers,verbs=update
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

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

	if capsule.GetGeneration() == capsule.Status.ObservedGeneration {
		return ctrl.Result{}, nil
	}

	status := &rigdevv1alpha1.CapsuleStatus{}
	var stepErrs []error
	for _, sf := range r.reconcileSteps {
		if err := sf(ctx, req, log, capsule, status); err != nil {
			stepErrs = append(stepErrs, err)
		}
	}

	if len(stepErrs) == 0 {
		status.ObservedGeneration = capsule.GetGeneration()
	}

	capsule.Status = *status
	if err := r.Status().Update(ctx, capsule); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, errors.Join(stepErrs...)
}

func (r *CapsuleReconciler) reconcileDeployment(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
	status *rigdevv1alpha1.CapsuleStatus,
) error {
	deploy, err := createDeployment(capsule, r.Scheme)
	if err != nil {
		status.Deployment.State = "failed"
		status.Deployment.Message = err.Error()
		return err
	}

	existingDeploy := &appsv1.Deployment{}
	if err = r.Get(ctx, req.NamespacedName, existingDeploy); err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("creating deployment")
			if err := r.Create(ctx, deploy); err != nil {
				status.Deployment.State = "failed"
				status.Deployment.Message = err.Error()
				return fmt.Errorf("could not create deployment: %w", err)
			}
			existingDeploy = deploy
		} else {
			status.Deployment.State = "failed"
			status.Deployment.Message = err.Error()
			return fmt.Errorf("could not fetch deployment: %w", err)
		}
	}

	// Edge case, this property is not carried over by k8s.
	delete(existingDeploy.Spec.Template.Annotations, "kubectl.kubernetes.io/restartedAt")

	err = upsertIfNewer(ctx, r, existingDeploy, deploy, log, capsule, status, func(t1, t2 *appsv1.Deployment) bool {
		return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
	})
	if err != nil {
		status.Deployment.State = "failed"
		status.Deployment.Message = err.Error()
	}
	return err
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
							Resources:    makeResourceRequirements(capsule),
						},
					},
					ServiceAccountName: capsule.Spec.ServiceAccountName,
					Volumes:            volumes,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(capsule, d, scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on deployment: %w", err)
	}

	return d, nil
}

func makeResourceRequirements(capsule *rigdevv1alpha1.Capsule) v1.ResourceRequirements {
	requests := utils.DefaultResources.Requests
	res := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    *resource.NewMilliQuantity(int64(requests.CpuMillis), resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(int64(requests.MemoryBytes), resource.DecimalSI),
		},
		Limits: v1.ResourceList{},
	}

	if capsule.Spec.Resources == nil {
		return res
	}
	for name, q := range capsule.Spec.Resources.Requests {
		res.Requests[name] = q
	}
	for name, q := range capsule.Spec.Resources.Limits {
		res.Limits[name] = q
	}

	return res
}

func (r *CapsuleReconciler) reconcileService(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
	status *rigdevv1alpha1.CapsuleStatus,
) error {
	service, err := createService(capsule, r.Scheme)
	if err != nil {
		return err
	}

	existingService := &v1.Service{}
	if err := r.Get(ctx, req.NamespacedName, existingService); err != nil {
		if kerrors.IsNotFound(err) {
			if len(capsule.Spec.Interfaces) == 0 {
				return nil
			}

			log.Info("creating service")
			if err := r.Create(ctx, service); err != nil {
				return fmt.Errorf("could not create service: %w", err)
			}
			existingService = service
		} else {
			return fmt.Errorf("could not fetch service: %w", err)
		}
	}

	if !IsOwnedBy(capsule, existingService) {
		if len(capsule.Spec.Interfaces) == 0 {
			log.Info("Found existing service not owned by capsule. Will not delete it.")
		} else {
			log.Info("Found existing service not owned by capsule. Will not update it.")
			return errors.New("found existing service not owned by capsule")
		}
	} else {
		if len(capsule.Spec.Interfaces) == 0 {
			log.Info("deleting service")
			if err := r.Delete(ctx, existingService); err != nil {
				return fmt.Errorf("could not delete service: %w", err)
			}
		} else {
			return upsertIfNewer(ctx, r, existingService, service, log, capsule, status, func(t1, t2 *v1.Service) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		}
	}

	return nil
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

func (r *CapsuleReconciler) reconcileCertificate(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
	status *rigdevv1alpha1.CapsuleStatus,
) error {
	crt, err := r.createCertificate(capsule, r.Scheme)
	if err != nil {
		return err
	}

	existingCrt := &cmv1.Certificate{}
	if err := r.Get(ctx, req.NamespacedName, existingCrt); err != nil {
		if kerrors.IsNotFound(err) {
			if !capsuleHasIngress(capsule) {
				return nil
			}
			if !r.ingressIsSupported() {
				log.V(1).Info("not creating certificate as ingress is not supported: cert-manager config missing")
				return nil
			}
			if !r.shouldCreateCertificateRessource() {
				log.V(1).Info("not creating certificate as operator is configured to use ingress annotations")
				return nil
			}

			log.Info("creating certificate")
			if err := r.Create(ctx, crt); err != nil {
				return fmt.Errorf("could not create certificate: %w", err)
			}
			existingCrt = crt
		} else {
			return fmt.Errorf("could not fetch certificate: %w", err)
		}
	}

	if !IsOwnedBy(capsule, existingCrt) {
		if capsuleHasIngress(capsule) {
			log.Info("Found existing certificate not owned by capsule. Will not update it.")
			return errors.New("found existing certificate not owned by capsule")
		} else {
			log.Info("Found existing certificate not owned by capsule. Will not delete it.")
		}
	} else {
		if r.ingressIsSupported() && r.shouldCreateCertificateRessource() && capsuleHasIngress(capsule) {
			return upsertIfNewer(ctx, r, existingCrt, crt, log, capsule, status, func(t1, t2 *cmv1.Certificate) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		} else {
			if !r.ingressIsSupported() {
				log.V(1).Info("deleting certificate as ingress is not supported: cert-manager config missing")
			} else if !r.shouldCreateCertificateRessource() {
				log.V(1).Info("deleting certificate becausee operator is configured to use ingress annotations")
			} else {
				log.Info("deleting certificate")
			}
			if err := r.Delete(ctx, existingCrt); err != nil {
				return fmt.Errorf("could not delete certificate: %w", err)
			}
		}
	}

	return nil
}

func (r *CapsuleReconciler) shouldCreateCertificateRessource() bool {
	return r.Config.Certmanager != nil &&
		r.Config.Certmanager.CreateCertificateResources
}

func (r *CapsuleReconciler) createCertificate(
	capsule *rigdevv1alpha1.Capsule,
	scheme *runtime.Scheme,
) (*cmv1.Certificate, error) {
	crt := &cmv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      capsule.Name,
			Namespace: capsule.Namespace,
		},
		Spec: cmv1.CertificateSpec{
			SecretName: fmt.Sprintf("%s-tls", capsule.Name),
		},
	}

	if r.Config.Certmanager != nil {
		crt.Spec.IssuerRef = cmmetav1.ObjectReference{
			Kind: cmv1.ClusterIssuerKind,
			Name: r.Config.Certmanager.ClusterIssuer,
		}
	}

	for _, inf := range capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			crt.Spec.DNSNames = append(crt.Spec.DNSNames, inf.Public.Ingress.Host)
		}
	}

	if err := controllerutil.SetControllerReference(capsule, crt, scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on certificate: %w", err)
	}

	return crt, nil
}

func (r *CapsuleReconciler) ingressIsSupported() bool {
	cm := r.Config.Certmanager
	return cm != nil && cm.ClusterIssuer != ""
}

func (r *CapsuleReconciler) reconcileIngress(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
	status *rigdevv1alpha1.CapsuleStatus,
) error {
	ing, err := r.createIngress(capsule, r.Scheme)
	if err != nil {
		return err
	}

	existingIng := &netv1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, existingIng); err != nil {
		if kerrors.IsNotFound(err) {
			if !capsuleHasIngress(capsule) {
				return nil
			}
			if !r.ingressIsSupported() {
				log.V(1).Info("ingress not supported: cert-manager config missing")
				return nil
			}

			log.Info("creating ingress")
			if err := r.Create(ctx, ing); err != nil {
				return fmt.Errorf("could not create ingress: %w", err)
			}
			existingIng = ing
		} else {
			return fmt.Errorf("could not fetch ingress: %w", err)
		}
	}

	if !IsOwnedBy(capsule, existingIng) {
		if capsuleHasIngress(capsule) {
			log.Info("Found existing ingress not owned by capsule. Will not update it.")
			return errors.New("found existing ingress not owned by capsule")
		} else {
			log.Info("Found existing ingress not owned by capsule. Will not delete it.")
		}
	} else {
		if r.ingressIsSupported() && capsuleHasIngress(capsule) {
			return upsertIfNewer(ctx, r, existingIng, ing, log, capsule, status, func(t1, t2 *netv1.Ingress) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		} else {
			if !r.ingressIsSupported() {
				log.V(1).Info("ingress not supported: cert-manager config missing")
			}
			log.Info("deleting ingress")
			if err := r.Delete(ctx, existingIng); err != nil {
				return fmt.Errorf("could not delete ingress: %w", err)
			}
		}
	}

	return nil
}

func capsuleHasIngress(capsule *rigdevv1alpha1.Capsule) bool {
	for _, inf := range capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			return true
		}
	}
	return false
}

func (r *CapsuleReconciler) createIngress(
	capsule *rigdevv1alpha1.Capsule,
	scheme *runtime.Scheme,
) (*netv1.Ingress, error) {
	ing := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        capsule.Name,
			Namespace:   capsule.Namespace,
			Annotations: r.Config.Ingress.Annotations,
		},
	}

	if r.Config.Ingress.ClassName != "" {
		ing.Spec.IngressClassName = ptr.New(r.Config.Ingress.ClassName)
	}

	if r.ingressIsSupported() && !r.shouldCreateCertificateRessource() {
		ing.Annotations["cert-manager.io/cluster-issuer"] = r.Config.Certmanager.ClusterIssuer
	}

	for _, inf := range capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
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
			if len(ing.Spec.TLS) == 0 {
				ing.Spec.TLS = []netv1.IngressTLS{{
					SecretName: fmt.Sprintf("%s-tls", capsule.Name),
				}}
			}
			ing.Spec.TLS[0].Hosts = append(ing.Spec.TLS[0].Hosts, inf.Public.Ingress.Host)
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
	status *rigdevv1alpha1.CapsuleStatus,
) error {
	svc, err := createLoadBalancer(capsule, r.Scheme)
	if err != nil {
		return err
	}

	nsName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-lb", req.NamespacedName.Name),
		Namespace: req.NamespacedName.Namespace,
	}
	existingSvc := &v1.Service{}
	if err := r.Get(ctx, nsName, existingSvc); err != nil {
		if kerrors.IsNotFound(err) {
			if !capsuleHasLoadBalancer(capsule) {
				return nil
			}

			log.Info("creating loadbalancer service")
			if err := r.Create(ctx, svc); err != nil {
				return fmt.Errorf("could not create loadbalancer: %w", err)
			}
			existingSvc = svc
		} else {
			return fmt.Errorf("could not fetch loadbalancer: %w", err)
		}
	}

	if !IsOwnedBy(capsule, existingSvc) {
		if capsuleHasLoadBalancer(capsule) {
			log.Info("Found existing loadbalancer service not owned by capsule. Will not update it.")
			return errors.New("found existing loadbalancer service not owned by capsule")
		} else {
			log.Info("Found existing loadbalancer service not owned by capsule. Will not delete it.")
		}
	} else {
		if capsuleHasLoadBalancer(capsule) {
			return upsertIfNewer(ctx, r, existingSvc, svc, log, capsule, status, func(t1, t2 *v1.Service) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		} else {
			log.Info("deleting loadbalancer service")
			if err := r.Delete(ctx, existingSvc); err != nil {
				return fmt.Errorf("could not delete loadbalancer service: %w", err)
			}
		}
	}

	return nil
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
				NodePort:   inf.Public.LoadBalancer.NodePort,
			})
		}
	}

	if err := controllerutil.SetControllerReference(capsule, svc, scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on loadbalancer service: %w", err)
	}

	return svc, nil
}

func (r *CapsuleReconciler) reconcileHorizontalPodAutoscaler(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *rigdevv1alpha1.Capsule,
	status *rigdevv1alpha1.CapsuleStatus,
) error {
	hpa, shouldHaveHPA, err := createHPA(capsule, r.Scheme)
	if err != nil {
		return err
	}
	existingHPA := &autoscalingv2.HorizontalPodAutoscaler{}
	hasExistingHPA := false
	if err = r.Get(ctx, client.ObjectKeyFromObject(hpa), existingHPA); err != nil {
		if kerrors.IsNotFound(err) && shouldHaveHPA {
			log.Info("creating horizontal pod autoscaler")
			if err := r.Create(ctx, hpa); err != nil {
				return fmt.Errorf("could not create horizontal pod autoscaler: %w", err)
			}
			existingHPA = hpa
		} else if err != nil {
			return fmt.Errorf("could not fetch horizontal pod autoscaler: %w", err)
		}
	} else {
		hasExistingHPA = true
	}
	if !shouldHaveHPA && hasExistingHPA {
		if err := r.Delete(ctx, existingHPA); err != nil {
			return err
		}
	}

	if shouldHaveHPA {
		return upsertIfNewer(ctx, r, existingHPA, hpa, log, capsule, status, func(t1, t2 *autoscalingv2.HorizontalPodAutoscaler) bool {
			return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
		})
	}

	return nil
}

func createHPA(capsule *rigdevv1alpha1.Capsule, scheme *runtime.Scheme) (*autoscalingv2.HorizontalPodAutoscaler, bool, error) {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-autoscaler", capsule.Name),
			Namespace: capsule.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       capsule.Name,
				APIVersion: appsv1.SchemeGroupVersion.Version,
			},
		},
	}
	if err := controllerutil.SetControllerReference(capsule, hpa, scheme); err != nil {
		return nil, true, err
	}

	scale := capsule.Spec.HorizontalScale
	var maxReplicas uint32
	var minReplicas uint32
	if scale.MinReplicas == nil {
		minReplicas = 1
	} else {
		minReplicas = *scale.MinReplicas
	}
	if scale.MaxReplicas == nil {
		maxReplicas = minReplicas
	} else {
		maxReplicas = *scale.MaxReplicas
	}
	if maxReplicas == 0 && minReplicas == 0 {
		capsule.Spec.Replicas = ptr.New(int32(0))
		return hpa, false, nil
	}

	if scale.CPUTarget != (rigdevv1alpha1.CPUTarget{}) {
		hpa.Spec.Metrics = append(hpa.Spec.Metrics, autoscalingv2.MetricSpec{
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

	hpa.Spec.MaxReplicas = int32(maxReplicas)
	hpa.Spec.MinReplicas = ptr.New(int32(minReplicas))

	return hpa, true, nil
}

func upsertIfNewer[T client.Object](ctx context.Context, r *CapsuleReconciler, current T, new T, log logr.Logger, capsule *rigdevv1alpha1.Capsule, status *rigdevv1alpha1.CapsuleStatus, equal func(t1 T, t2 T) bool) error {
	kind := current.GetObjectKind().GroupVersionKind().Kind

	res := rigdevv1alpha1.OwnedResource{
		Ref: &v1.TypedLocalObjectReference{
			Kind: kind,
			Name: new.GetName(),
		},
		State: "created",
	}
	defer func() {
		status.OwnedResources = append(status.OwnedResources, res)
	}()

	if !IsOwnedBy(capsule, new) {
		log.Info("Found existing resource not owned by capsule. Will not update it.", "kind", kind, "name", new.GetName())
		res.State = "failed"
		res.Message = "found existing resource not owned by capsule"
		return fmt.Errorf("found existing %s not owned by capsule", kind)
	}

	// Dry run to fully materialize the new spec.
	if err := r.Update(ctx, new, client.DryRunAll); err != nil {
		res.State = "failed"
		res.Message = err.Error()
		return fmt.Errorf("could not update %s: %w", kind, err)
	}

	if !equal(new, current) {
		log.Info("updating", "kind", kind)
		if err := r.Update(ctx, new); err != nil {
			res.State = "failed"
			res.Message = err.Error()
			return fmt.Errorf("could not update %s: %w", kind, err)
		}
	}

	return nil
}
