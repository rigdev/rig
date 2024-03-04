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

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/go-logr/logr"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/service/capabilities"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// CapsuleReconciler reconciles a Capsule object
type CapsuleReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	Config              *configv1alpha1.OperatorConfig
	ClientSet           clientset.Interface
	CapabilitiesService capabilities.Service
	Pipeline            *pipeline.Pipeline
}

const (
	AnnotationChecksumFiles     = "rig.dev/config-checksum-files"
	AnnotationChecksumAutoEnv   = "rig.dev/config-checksum-auto-env"
	AnnotationChecksumEnv       = "rig.dev/config-checksum-env"
	AnnotationChecksumSharedEnv = "rig.dev/config-checksum-shared-env"

	LabelSharedConfig = "rig.dev/shared-config"
	LabelCapsule      = "rig.dev/capsule"
	LabelCron         = "batch.kubernets.io/cronjob"

	fieldFilesConfigMapName = ".spec.files.configMap.name"
	fieldFilesSecretName    = ".spec.files.secret.name"
	fieldEnvConfigMapName   = ".spec.env.from.configMapName"
	fieldEnvSecretName      = ".spec.env.from.secretName"
)

// SetupWithManager sets up the controller with the Manager.
func (r *CapsuleReconciler) SetupWithManager(mgr ctrl.Manager, logger logr.Logger) error {
	// TODO Where to get the context from?
	ctx := context.Background()

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

	steps, err := GetDefaultPipelineSteps(ctx, r.CapabilitiesService, r.Config)
	if err != nil {
		return err
	}

	r.Pipeline = pipeline.New(r.Client, r.Config, r.Scheme, logger)
	for _, step := range steps {
		r.Pipeline.AddStep(step)
	}

	for _, step := range r.Config.Steps {
		ps, err := plugin.NewStep(step, logger)
		if err != nil {
			return err
		}

		r.Pipeline.AddStep(ps)
	}

	configEventHandler := handler.EnqueueRequestsFromMapFunc(findCapsulesForConfig(mgr))

	b := ctrl.NewControllerManagedBy(mgr)
	if capabilities.GetHasPrometheusServiceMonitor() {
		b = b.Owns(&monitorv1.ServiceMonitor{})
	}

	if r.Config.VerticalPodAutoscaler.Enabled {
		b = b.Owns(&vpav1.VerticalPodAutoscaler{})
	}

	return b.
		For(&v1alpha2.Capsule{}).
		Owns(&appsv1.Deployment{}).
		Owns(&v1.Service{}).
		Owns(&netv1.Ingress{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&cmv1.Certificate{}).
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
		if sharedConfig := o.GetLabels()[LabelSharedConfig]; sharedConfig == "true" {
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
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("could not fetch Capsule: %w", err)
	}

	if _, err := r.Pipeline.RunCapsule(ctx, capsule); err != nil {
		log.Error(err, "reconciliation ended with error")
		return ctrl.Result{}, err
	}

	log.Info("reconciliation completed successfully")

	return ctrl.Result{}, nil
}

func GetDefaultPipelineSteps(ctx context.Context, capSvc capabilities.Service, cfg *v1alpha1.OperatorConfig) ([]pipeline.Step, error) {
	capabilities, err := capSvc.Get(ctx)
	if err != nil {
		return nil, err
	}

	var steps []pipeline.Step

	steps = append(steps,
		NewServiceAccountStep(),
		NewDeploymentStep(),
		NewVPAStep(cfg),
		NewNetworkStep(cfg),
		NewCronJobStep(),
	)

	if capabilities.GetHasPrometheusServiceMonitor() {
		steps = append(steps, NewServiceMonitorStep(cfg))
	}

	return steps, nil
}
