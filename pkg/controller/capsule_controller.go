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
	"fmt"
	"strconv"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/objectstatus"
	svc_pipeline "github.com/rigdev/rig/pkg/service/pipeline"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// CapsuleReconciler reconciles a Capsule object
type CapsuleReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	Config              *configv1alpha1.OperatorConfig
	CapabilitiesService capabilities.Service
	PipelineService     svc_pipeline.Service
	ObjectStatusService objectstatus.Service
}

const (
	CleanupFinalizer = "rig.dev/capsule-cleanup"

	fieldFilesConfigMapName = ".spec.files.configMap.name"
	fieldFilesSecretName    = ".spec.files.secret.name"
	fieldEnvConfigMapName   = ".spec.env.from.configMapName"
	fieldEnvSecretName      = ".spec.env.from.secretName"
)

// SetupWithManager sets up the controller with the Manager.
func (r *CapsuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.TODO()

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1alpha2.Capsule{},
		fieldFilesConfigMapName,
		func(o client.Object) []string {
			capsule := o.(*v1alpha2.Capsule)
			var cms []string
			for _, f := range capsule.Spec.Files {
				if f.Ref != nil && f.Ref.Kind == "ConfigMap" {
					cms = append(cms, f.Ref.Name)
				}
			}
			return cms
		},
	); err != nil {
		return fmt.Errorf("could not setup indexer for %s: %w", fieldFilesConfigMapName, err)
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1alpha2.Capsule{},
		fieldFilesSecretName,
		func(o client.Object) []string {
			capsule := o.(*v1alpha2.Capsule)
			var ss []string
			for _, f := range capsule.Spec.Files {
				if f.Ref != nil && f.Ref.Kind == "Secret" {
					ss = append(ss, f.Ref.Name)
				}
			}
			return ss
		},
	); err != nil {
		return fmt.Errorf("could not setup indexer for %s: %w", fieldFilesSecretName, err)
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1alpha2.Capsule{},
		fieldEnvConfigMapName,
		func(o client.Object) []string {
			capsule := o.(*v1alpha2.Capsule)
			var cms []string
			for _, from := range capsule.Spec.Env.From {
				if from.Kind == "ConfigMap" {
					cms = append(cms, from.Name)
				}
			}
			return cms
		},
	); err != nil {
		return fmt.Errorf("could not setup indexer for %s: %w", fieldEnvConfigMapName, err)
	}

	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1alpha2.Capsule{},
		fieldEnvSecretName,
		func(o client.Object) []string {
			capsule := o.(*v1alpha2.Capsule)
			var ss []string
			for _, from := range capsule.Spec.Env.From {
				if from.Kind == "Secret" {
					ss = append(ss, from.Name)
				}
			}
			return ss
		},
	); err != nil {
		return fmt.Errorf("could not setup indexer for %s: %w", fieldEnvSecretName, err)
	}

	capabilities, err := r.CapabilitiesService.Get(ctx)
	if err != nil {
		return err
	}

	configEventHandler := handler.EnqueueRequestsFromMapFunc(findCapsulesForConfig(mgr))

	b := ctrl.NewControllerManagedBy(mgr)
	if capabilities.GetHasPrometheusServiceMonitor() {
		b = b.Owns(&monitorv1.ServiceMonitor{})
	}

	if r.Config.Pipeline.VPAStep.Plugin != "" {
		b = b.Owns(&vpav1.VerticalPodAutoscaler{})
	}

	if capabilities.GetIngress() {
		b = b.Owns(&cmv1.Certificate{})
	}

	return b.
		For(&v1alpha2.Capsule{}).
		Owns(&appsv1.Deployment{}).
		Owns(&v1.Service{}).
		Owns(&netv1.Ingress{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&batchv1.CronJob{}).
		Watches(
			&v1.ConfigMap{},
			configEventHandler,
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&v1.Secret{},
			configEventHandler,
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func findCapsulesForConfig(mgr ctrl.Manager) handler.MapFunc {
	scheme := mgr.GetScheme()
	log := mgr.GetLogger().WithName("configEventHandler")
	c := mgr.GetClient()

	return func(ctx context.Context, o client.Object) []ctrl.Request {
		var capsulesWithReference v1alpha2.CapsuleList
		// Queue reconcile for all capsules in namespace if this is a shared config
		if sharedConfig := o.GetLabels()[pipeline.LabelSharedConfig]; sharedConfig == "true" {
			if err := c.List(ctx, &capsulesWithReference, client.InNamespace(o.GetNamespace())); err != nil {
				log.Error(err, "could not get capsules")
			}
			requests := make([]ctrl.Request, len(capsulesWithReference.Items))
			for i, c := range capsulesWithReference.Items {
				requests[i] = ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&c)}
			}
			return requests
		}

		gvks, _, err := scheme.ObjectKinds(o)
		if err != nil {
			log.Error(err, "could not get ObjectKinds from object")
			return nil
		}

		gvk := gvks[0]
		log := log.WithValues(gvk.Kind, o)

		var (
			filesRefField string
			envRefField   string
		)
		switch gvk.Kind {
		case "Secret":
			filesRefField = fieldFilesSecretName
			envRefField = fieldEnvSecretName
		case "ConfigMap":
			filesRefField = fieldFilesConfigMapName
			envRefField = fieldEnvConfigMapName
		default:
			log.Error(fmt.Errorf("unsupported Kind: %s", gvk.Kind), "unsupported kind")
			return nil
		}

		var requests []ctrl.Request

		// Queue reconcile for all capsules referencing this config
		if err = c.List(ctx, &capsulesWithReference, &client.ListOptions{
			Namespace:     o.GetNamespace(),
			FieldSelector: fields.SelectorFromSet(fields.Set{filesRefField: o.GetName()}),
		}); err != nil {
			log.Error(err, "could not list capsules with reference to object", "err", fmt.Sprintf("%+v\n", err))
			return nil
		}
		for _, capsule := range capsulesWithReference.Items {
			requests = append(requests, ctrl.Request{
				NamespacedName: client.ObjectKeyFromObject(&capsule),
			})
		}

		// Queue reconcile for automatic env
		var capsule v1alpha2.Capsule
		err = c.Get(ctx, client.ObjectKeyFromObject(o), &capsule)
		if err != nil && !kerrors.IsNotFound(err) {
			log.Error(err, "could not get capsule for object")
			return nil
		}
		if err == nil {
			if !capsule.Spec.Env.DisableAutomatic {
				requests = append(requests, ctrl.Request{
					NamespacedName: client.ObjectKeyFromObject(o),
				})
			}
		}

		// Queue reconcile for specific env
		if err = c.List(ctx, &capsulesWithReference, &client.ListOptions{
			Namespace:     o.GetNamespace(),
			FieldSelector: fields.SelectorFromSet(fields.Set{envRefField: o.GetName()}),
		}); err != nil {
			log.Error(err, "could not list capsules with reference to object", "err", fmt.Sprintf("%+v\n", err))
			return nil
		}
		for _, capsule := range capsulesWithReference.Items {
			requests = append(requests, ctrl.Request{
				NamespacedName: client.ObjectKeyFromObject(&capsule),
			})
		}

		return requests
	}
}

//+kubebuilder:rbac:groups=rig.dev,resources=capsules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rig.dev,resources=capsules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rig.dev,resources=capsules/finalizers,verbs=update
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps;secrets,verbs=get;list;watch

// Reconcile compares the state specified by the Capsule object against the
// actual cluster state, and then performs operations to make the cluster state
// reflect the state specified by the Capsule.
func (r *CapsuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO: use rig logger
	log := log.FromContext(ctx)
	log.Info("reconciliation started")

	capsule := &v1alpha2.Capsule{}
	if err := r.Get(ctx, req.NamespacedName, capsule); err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("capsule not found, aborting")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("could not fetch Capsule: %w", err)
	}

	// Test for deletion marker.
	if capsule.ObjectMeta.DeletionTimestamp.IsZero() {
		// Not for deletion. Add finalizer if missing.
		if controllerutil.AddFinalizer(capsule, CleanupFinalizer) {
			if err := r.Update(ctx, capsule); err != nil {
				log.Error(err, "error adding finalizer")
				return ctrl.Result{}, nil
			}

			return ctrl.Result{Requeue: true}, nil
		}
	} else {
		log.Info("capsule should be deleted")
		if _, err := r.PipelineService.GetDefaultPipeline().DeleteCapsule(ctx, capsule, r.Client); err != nil {
			log.Error(err, "delete ended with error")
			return ctrl.Result{}, err
		}

		r.ObjectStatusService.UnregisterCapsule(capsule.GetNamespace(), capsule.GetName())

		// Remove finalizer if present.
		if controllerutil.RemoveFinalizer(capsule, CleanupFinalizer) {
			if err := r.Update(ctx, capsule); err != nil {
				log.Error(err, "error removing finalizer")
				return ctrl.Result{}, err
			}
		}

		log.Info("capsule is deleted")
		return ctrl.Result{}, nil
	}

	r.ObjectStatusService.RegisterCapsule(capsule.GetNamespace(), capsule.GetName())

	var options []pipeline.CapsuleRequestOption
	if v, _ := strconv.ParseBool(capsule.Annotations[pipeline.AnnotationOverrideOwnership]); v {
		options = append(options, pipeline.WithForce())
	}

	if _, err := r.PipelineService.GetDefaultPipeline().RunCapsule(ctx, capsule, r.Client, options...); err != nil {
		log.Error(err, "reconciliation ended with error")
		return ctrl.Result{}, err
	}

	log.Info("reconciliation completed successfully")

	return ctrl.Result{}, nil
}
