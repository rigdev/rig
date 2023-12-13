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
	"crypto/sha256"
	"errors"
	"fmt"
	"net/url"
	"path"
	"slices"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/go-logr/logr"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/utils"
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/equality"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	Scheme    *runtime.Scheme
	Config    *configv1alpha1.OperatorConfig
	ClientSet clientset.Interface

	reconcileSteps []reconcileStepFunc
}

type reconcileStepFunc func(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
) error

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
func (r *CapsuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
			if capsule.Spec.Env == nil {
				return nil
			}
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
			if capsule.Spec.Env == nil {
				return nil
			}
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

	crds, err := r.ClientSet.ApiextensionsV1().CustomResourceDefinitions().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	hasServiceMonitor := false
	for _, crd := range crds.Items {
		if crd.Name == "servicemonitors.monitoring.coreos.com" {
			hasServiceMonitor = true
			break
		}
	}

	r.reconcileSteps = []reconcileStepFunc{
		r.reconcileHorizontalPodAutoscaler,
		r.reconcileDeployment,
		r.reconcileService,
		r.reconcileCertificate,
		r.reconcileIngress,
		r.reconcileLoadBalancer,
		r.reconcileServiceAccount,
		r.reconcileCronJobs,
	}

	configEventHandler := handler.EnqueueRequestsFromMapFunc(findCapsulesForConfig(mgr))

	b := ctrl.NewControllerManagedBy(mgr)
	if hasServiceMonitor {
		r.reconcileSteps = append(r.reconcileSteps, r.reconcilePrometheusServiceMonitor)
		b = b.Owns(&monitorv1.ServiceMonitor{})
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
			if capsule.Spec.Env == nil || !capsule.Spec.Env.DisableAutomatic {
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

	status := &v1alpha2.CapsuleStatus{
		Deployment: &v1alpha2.DeploymentStatus{},
	}
	var stepErrs []error
	for _, sf := range r.reconcileSteps {
		if err := sf(ctx, req, log, capsule, status); err != nil {
			stepErrs = append(stepErrs, err)
		}
	}

	if len(stepErrs) == 0 {
		status.ObservedGeneration = capsule.GetGeneration()
	} else {
		var errs []string
		for _, e := range stepErrs {
			errs = append(errs, e.Error())
		}
		status.Errors = errs
	}

	capsule.Status = status
	if err := r.Status().Update(ctx, capsule); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, errors.Join(stepErrs...)
}

type configs struct {
	configMaps          map[string]*v1.ConfigMap
	secrets             map[string]*v1.Secret
	sharedEnvConfigMaps []string
	sharedEnvSecrets    []string
}

func (c *configs) hasSharedConfig() bool {
	return len(c.sharedEnvConfigMaps) > 0 || len(c.sharedEnvSecrets) > 0
}

type checksums struct {
	sharedEnv string
	autoEnv   string
	env       string
	files     string
}

func (r *CapsuleReconciler) configChecksums(
	capsule *v1alpha2.Capsule,
	configs *configs,
) (*checksums, error) {
	sharedEnv, err := r.configSharedEnvChecksum(configs)
	if err != nil {
		return nil, err
	}

	autoEnv, err := r.configAutoEnvChecksum(
		configs.configMaps[capsule.GetName()],
		configs.secrets[capsule.GetName()],
	)
	if err != nil {
		return nil, err
	}

	env, err := r.configEnvChecksum(capsule, configs)
	if err != nil {
		return nil, err
	}

	files, err := r.configFilesChecksum(capsule, configs)
	if err != nil {
		return nil, err
	}

	return &checksums{
		sharedEnv: sharedEnv,
		autoEnv:   autoEnv,
		env:       env,
		files:     files,
	}, nil
}

func (r *CapsuleReconciler) configSharedEnvChecksum(
	configs *configs,
) (string, error) {
	if !configs.hasSharedConfig() {
		return "", nil
	}

	h := sha256.New()

	configMaps := slices.Clone(configs.sharedEnvConfigMaps)
	slices.Sort(configMaps)
	secrets := slices.Clone(configs.sharedEnvSecrets)
	slices.Sort(secrets)

	for _, name := range configMaps {
		if err := hash.ConfigMap(h, configs.configMaps[name]); err != nil {
			return "", err
		}
	}
	for _, name := range secrets {
		if err := hash.Secret(h, configs.secrets[name]); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *CapsuleReconciler) configAutoEnvChecksum(
	configMap *v1.ConfigMap,
	secret *v1.Secret,
) (string, error) {
	if configMap == nil && secret == nil {
		return "", nil
	}

	h := sha256.New()

	if configMap != nil {
		if err := hash.ConfigMap(h, configMap); err != nil {
			return "", err
		}
	}
	if secret != nil {
		if err := hash.Secret(h, secret); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *CapsuleReconciler) configEnvChecksum(
	capsule *v1alpha2.Capsule,
	configs *configs,
) (string, error) {
	if capsule.Spec.Env == nil || len(capsule.Spec.Env.From) == 0 {
		return "", nil
	}

	h := sha256.New()
	for _, e := range capsule.Spec.Env.From {
		switch e.Kind {
		case "ConfigMap":
			if err := hash.ConfigMap(h, configs.configMaps[e.Name]); err != nil {
				return "", err
			}
		case "Secret":
			if err := hash.Secret(h, configs.secrets[e.Name]); err != nil {
				return "", err
			}
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *CapsuleReconciler) configFilesChecksum(
	capsule *v1alpha2.Capsule,
	configs *configs,
) (string, error) {
	if len(capsule.Spec.Files) == 0 {
		return "", nil
	}

	referencedKeysBySecretName := map[string]map[string]struct{}{}
	referencedKeysByConfigMapName := map[string]map[string]struct{}{}
	for _, f := range capsule.Spec.Files {
		switch f.Ref.Kind {
		case "ConfigMap":
			if _, ok := referencedKeysByConfigMapName[f.Ref.Name]; ok {
				referencedKeysByConfigMapName[f.Ref.Name][f.Ref.Key] = struct{}{}
				continue
			}
			referencedKeysByConfigMapName[f.Ref.Name] = map[string]struct{}{
				f.Ref.Key: {},
			}
		case "Secret":
			if _, ok := referencedKeysBySecretName[f.Ref.Name]; ok {
				referencedKeysBySecretName[f.Ref.Name][f.Ref.Key] = struct{}{}
				continue
			}
			referencedKeysBySecretName[f.Ref.Name] = map[string]struct{}{
				f.Ref.Key: {},
			}
		}
	}

	secretNames := maps.Keys(referencedKeysBySecretName)
	slices.Sort(secretNames)
	configMapNames := maps.Keys(referencedKeysByConfigMapName)
	slices.Sort(configMapNames)
	h := sha256.New()
	for _, name := range secretNames {
		if err := hash.SecretKeys(
			h,
			maps.Keys(referencedKeysBySecretName[name]),
			configs.secrets[name],
		); err != nil {
			return "", err
		}
	}
	for _, name := range configMapNames {
		if err := hash.ConfigMapKeys(
			h,
			maps.Keys(referencedKeysByConfigMapName[name]),
			configs.configMaps[name],
		); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *CapsuleReconciler) getConfigs(
	ctx context.Context,
	req ctrl.Request,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
) (*configs, error) {
	cfgs := &configs{
		configMaps: map[string]*v1.ConfigMap{},
		secrets:    map[string]*v1.Secret{},
	}

	// Get shared env
	var configMapList v1.ConfigMapList
	if err := r.Client.List(ctx, &configMapList, &client.ListOptions{
		Namespace: req.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			LabelSharedConfig: "true",
		}),
	}); err != nil {
		return nil, fmt.Errorf("could not list shared env configmaps: %w", err)
	}
	cfgs.sharedEnvConfigMaps = make([]string, len(configMapList.Items))
	for i, cm := range configMapList.Items {
		cfgs.sharedEnvConfigMaps[i] = cm.GetName()
		cfgs.configMaps[cm.Name] = &cm
	}
	var secretList v1.SecretList
	if err := r.Client.List(ctx, &secretList, &client.ListOptions{
		Namespace: req.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			LabelSharedConfig: "true",
		}),
	}); err != nil {
		return nil, fmt.Errorf("could not list shared env secrets: %w", err)
	}
	cfgs.sharedEnvSecrets = make([]string, len(secretList.Items))
	for i, s := range secretList.Items {
		cfgs.sharedEnvSecrets[i] = s.GetName()
		cfgs.secrets[s.Name] = &s
	}

	env := capsule.Spec.Env
	if env == nil {
		env = &v1alpha2.Env{}
	}

	// Get automatic env
	if !env.DisableAutomatic {
		if err := r.getUsedSource(ctx, capsule, status, cfgs, "ConfigMap", req.NamespacedName.Name, false); err != nil {
			return nil, err
		}

		if err := r.getUsedSource(ctx, capsule, status, cfgs, "Secret", req.NamespacedName.Name, false); err != nil {
			return nil, err
		}
	}

	// Get envs
	for _, e := range env.From {
		if err := r.getUsedSource(ctx, capsule, status, cfgs, e.Kind, e.Name, true); err != nil {
			return nil, err
		}
	}

	// Get files
	for _, f := range capsule.Spec.Files {
		if err := r.getUsedSource(ctx, capsule, status, cfgs, f.Ref.Kind, f.Ref.Name, true); err != nil {
			return nil, err
		}
	}

	return cfgs, nil
}

func (r *CapsuleReconciler) getUsedSource(
	ctx context.Context,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
	cfgs *configs,
	kind string,
	name string,
	required bool,
) (err error) {
	ref := v1alpha2.UsedResource{
		Ref: &v1.TypedLocalObjectReference{
			Kind: kind,
			Name: name,
		},
	}

	defer func() {
		if kerrors.IsNotFound(err) && !required {
			ref.State = "missing"
			err = nil
		} else if err != nil {
			ref.State = "error"
			ref.Message = err.Error()
		} else {
			ref.State = "found"
		}

		status.UsedResources = append(status.UsedResources, ref)
	}()

	switch kind {
	case "ConfigMap":
		if _, ok := cfgs.configMaps[name]; ok {
			return nil
		}
		var cm v1.ConfigMap
		if err := r.Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: capsule.Namespace,
		}, &cm); err != nil {
			return fmt.Errorf("could not get referenced environment configmap: %w", err)
		}
		cfgs.configMaps[cm.Name] = &cm
	case "Secret":
		if _, ok := cfgs.secrets[name]; ok {
			return nil
		}
		var s v1.Secret
		if err := r.Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: capsule.Namespace,
		}, &s); err != nil {
			return fmt.Errorf("could not get referenced environment secret: %w", err)
		}
		cfgs.secrets[s.Name] = &s
	}

	return nil
}

func (r *CapsuleReconciler) reconcileDeployment(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
) error {
	cfgs, err := r.getConfigs(ctx, req, capsule, status)
	if err != nil {
		return err
	}

	checksums, err := r.configChecksums(capsule, cfgs)
	if err != nil {
		return err
	}

	existingDeploy := &appsv1.Deployment{}
	hasExistingDeployment := true
	if err = r.Get(ctx, req.NamespacedName, existingDeploy); err != nil {
		if kerrors.IsNotFound(err) {
			hasExistingDeployment = false
		} else {
			status.Deployment.State = "failed"
			status.Deployment.Message = err.Error()
			return fmt.Errorf("could not fetch deployment: %w", err)
		}
	}

	deploy, err := createDeployment(capsule, r.Scheme, cfgs, checksums, existingDeploy)
	if err != nil {
		return err
	}

	if !hasExistingDeployment {
		log.Info("creating deployment")
		if err := r.Create(ctx, deploy); err != nil {
			status.Deployment.State = "failed"
			status.Deployment.Message = err.Error()
			return fmt.Errorf("could not create deployment: %w", err)
		}
		existingDeploy = deploy
	}

	if err != nil {
		status.Deployment.State = "failed"
		status.Deployment.Message = err.Error()
		return err
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
	capsule *v1alpha2.Capsule,
	scheme *runtime.Scheme,
	configs *configs,
	checksums *checksums,
	existingDeployment *appsv1.Deployment,
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
		switch f.Ref.Kind {
		case "ConfigMap":
			name = "configmap-" + strings.ReplaceAll(f.Ref.Name, ".", "-")
			volumes = append(volumes, v1.Volume{
				Name: name,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: f.Ref.Name,
						},
						Items: []v1.KeyToPath{
							{
								Key:  f.Ref.Key,
								Path: path.Base(f.Path),
							},
						},
					},
				},
			})
		case "Secret":
			name = "secret-" + strings.ReplaceAll(f.Ref.Name, ".", "-")
			volumes = append(volumes, v1.Volume{
				Name: name,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: f.Ref.Name,
						Items: []v1.KeyToPath{
							{
								Key:  f.Ref.Key,
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

	podAnnotations := map[string]string{}
	maps.Copy(podAnnotations, capsule.Annotations)
	if checksums.files != "" {
		podAnnotations[AnnotationChecksumFiles] = checksums.files
	}
	if checksums.autoEnv != "" {
		podAnnotations[AnnotationChecksumAutoEnv] = checksums.autoEnv
	}
	if checksums.env != "" {
		podAnnotations[AnnotationChecksumEnv] = checksums.env
	}
	if checksums.sharedEnv != "" {
		podAnnotations[AnnotationChecksumSharedEnv] = checksums.sharedEnv
	}

	var envFrom []v1.EnvFromSource
	if capsule.Spec.Env == nil || !capsule.Spec.Env.DisableAutomatic {
		if _, ok := configs.configMaps[capsule.GetName()]; ok {
			envFrom = append(envFrom, v1.EnvFromSource{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: capsule.GetName()},
				},
			})
		}
		if _, ok := configs.secrets[capsule.GetName()]; ok {
			envFrom = append(envFrom, v1.EnvFromSource{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: capsule.GetName()},
				},
			})
		}
	}

	if capsule.Spec.Env != nil {
		for _, e := range capsule.Spec.Env.From {
			switch e.Kind {
			case "ConfigMap":
				envFrom = append(envFrom, v1.EnvFromSource{
					ConfigMapRef: &v1.ConfigMapEnvSource{
						LocalObjectReference: v1.LocalObjectReference{Name: e.Name},
					},
				})
			case "Secret":
				envFrom = append(envFrom, v1.EnvFromSource{
					SecretRef: &v1.SecretEnvSource{
						LocalObjectReference: v1.LocalObjectReference{Name: e.Name},
					},
				})
			}
		}
	}

	for _, name := range configs.sharedEnvConfigMaps {
		envFrom = append(envFrom, v1.EnvFromSource{
			ConfigMapRef: &v1.ConfigMapEnvSource{
				LocalObjectReference: v1.LocalObjectReference{Name: name},
			},
		})
	}
	for _, name := range configs.sharedEnvSecrets {
		envFrom = append(envFrom, v1.EnvFromSource{
			SecretRef: &v1.SecretEnvSource{
				LocalObjectReference: v1.LocalObjectReference{Name: name},
			},
		})
	}

	c := v1.Container{
		Name:         capsule.Name,
		Image:        capsule.Spec.Image,
		EnvFrom:      envFrom,
		VolumeMounts: volumeMounts,
		Ports:        ports,
		Resources:    makeResourceRequirements(capsule),
		Args:         capsule.Spec.Args,
	}

	if capsule.Spec.Command != "" {
		c.Command = []string{capsule.Spec.Command}
	}

	for _, i := range capsule.Spec.Interfaces {
		if i.Liveness != nil {
			c.LivenessProbe = &v1.Probe{
				ProbeHandler: v1.ProbeHandler{
					HTTPGet: &v1.HTTPGetAction{
						Path: i.Liveness.Path,
						Port: intstr.FromInt32(i.Port),
					},
				},
			}
		}
		if i.Readiness != nil {
			c.ReadinessProbe = &v1.Probe{
				ProbeHandler: v1.ProbeHandler{
					HTTPGet: &v1.HTTPGetAction{
						Path: i.Readiness.Path,
						Port: intstr.FromInt32(i.Port),
					},
				},
			}
		}
	}

	replicas := ptr.New(int32(capsule.Spec.Scale.Horizontal.Instances.Min))
	hasHPA, err := shouldCreateHPA(capsule, scheme)
	if err != nil {
		return nil, err
	}
	if hasHPA {
		if existingDeployment != nil && existingDeployment.Spec.Replicas != nil {
			replicas = ptr.New(*existingDeployment.Spec.Replicas)
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
					LabelCapsule: capsule.Name,
				},
			},
			Replicas: replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: podAnnotations,
					Labels: map[string]string{
						LabelCapsule: capsule.Name,
					},
				},
				Spec: v1.PodSpec{
					Containers:         []v1.Container{c},
					ServiceAccountName: capsule.Name,
					Volumes:            volumes,
					NodeSelector:       capsule.Spec.NodeSelector,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(capsule, d, scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on deployment: %w", err)
	}

	return d, nil
}

func makeResourceRequirements(capsule *v1alpha2.Capsule) v1.ResourceRequirements {
	requests := utils.DefaultResources.Requests
	res := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    *resource.NewMilliQuantity(int64(requests.CpuMillis), resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(int64(requests.MemoryBytes), resource.DecimalSI),
		},
		Limits: v1.ResourceList{},
	}

	if capsule.Spec.Scale.Vertical == nil {
		return res
	}
	if c := capsule.Spec.Scale.Vertical.CPU; c != nil {
		if c.Request != nil && !c.Request.IsZero() {
			res.Requests[v1.ResourceCPU] = *c.Request
		}
		if c.Limit != nil && !c.Limit.IsZero() {
			res.Limits[v1.ResourceCPU] = *c.Limit
		}
	}
	if m := capsule.Spec.Scale.Vertical.Memory; m != nil {
		if m.Request != nil && !m.Request.IsZero() {
			res.Requests[v1.ResourceMemory] = *m.Request
		}
		if m.Limit != nil && !m.Limit.IsZero() {
			res.Limits[v1.ResourceMemory] = *m.Limit
		}
	}
	if g := capsule.Spec.Scale.Vertical.GPU; g != nil && !g.Request.IsZero() {
		res.Requests["nvidia.com/gpu"] = g.Request
	}

	return res
}

func (r *CapsuleReconciler) reconcileService(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
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
	capsule *v1alpha2.Capsule,
	scheme *runtime.Scheme,
) (*v1.Service, error) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      capsule.Name,
			Namespace: capsule.Namespace,
			Labels: map[string]string{
				LabelCapsule: capsule.Name,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				LabelCapsule: capsule.Name,
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
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
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
		}
		log.Info("Found existing certificate not owned by capsule. Will not delete it.")
	} else {
		if r.ingressIsSupported() && r.shouldCreateCertificateRessource() && capsuleHasIngress(capsule) {
			return upsertIfNewer(ctx, r, existingCrt, crt, log, capsule, status, func(t1, t2 *cmv1.Certificate) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		}
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

	return nil
}

func (r *CapsuleReconciler) shouldCreateCertificateRessource() bool {
	return r.Config.Certmanager != nil &&
		r.Config.Certmanager.CreateCertificateResources
}

func (r *CapsuleReconciler) createCertificate(
	capsule *v1alpha2.Capsule,
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
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
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
		}
		log.Info("Found existing ingress not owned by capsule. Will not delete it.")
	} else {
		if r.ingressIsSupported() && capsuleHasIngress(capsule) {
			return upsertIfNewer(ctx, r, existingIng, ing, log, capsule, status, func(t1, t2 *netv1.Ingress) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		}
		if !r.ingressIsSupported() {
			log.V(1).Info("ingress not supported: cert-manager config missing")
		}
		log.Info("deleting ingress")
		if err := r.Delete(ctx, existingIng); err != nil {
			return fmt.Errorf("could not delete ingress: %w", err)
		}
	}

	return nil
}

func capsuleHasIngress(capsule *v1alpha2.Capsule) bool {
	for _, inf := range capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			return true
		}
	}
	return false
}

func (r *CapsuleReconciler) createIngress(
	capsule *v1alpha2.Capsule,
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
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
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
		}
		log.Info("Found existing loadbalancer service not owned by capsule. Will not delete it.")
	} else {
		if capsuleHasLoadBalancer(capsule) {
			return upsertIfNewer(ctx, r, existingSvc, svc, log, capsule, status, func(t1, t2 *v1.Service) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		}
		log.Info("deleting loadbalancer service")
		if err := r.Delete(ctx, existingSvc); err != nil {
			return fmt.Errorf("could not delete loadbalancer service: %w", err)
		}
	}

	return nil
}

func capsuleHasLoadBalancer(capsule *v1alpha2.Capsule) bool {
	for _, inf := range capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.LoadBalancer != nil {
			return true
		}
	}
	return false
}

func createLoadBalancer(
	capsule *v1alpha2.Capsule,
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
				LabelCapsule: capsule.Name,
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
		return nil, fmt.Errorf("could not set owner reference on loadbalancer service: %w", err)
	}

	return svc, nil
}

func (r *CapsuleReconciler) reconcileHorizontalPodAutoscaler(
	ctx context.Context,
	_ ctrl.Request,
	log logr.Logger,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
) error {
	hpa, shouldHaveHPA, err := createHPA(capsule, r.Scheme)
	if err != nil {
		return err
	}
	existingHPA := &autoscalingv2.HorizontalPodAutoscaler{}
	if err = r.Get(ctx, client.ObjectKeyFromObject(hpa), existingHPA); err != nil {
		if kerrors.IsNotFound(err) {
			if shouldHaveHPA {
				log.Info("creating horizontal pod autoscaler")
				if err := r.Create(ctx, hpa); err != nil {
					return fmt.Errorf("could not create horizontal pod autoscaler: %w", err)
				}
			}
			return nil
		} else if err != nil {
			return fmt.Errorf("could not fetch horizontal pod autoscaler: %w", err)
		}
	}

	if !shouldHaveHPA {
		if err := r.Delete(ctx, existingHPA); err != nil {
			return err
		}
	}

	return upsertIfNewer(
		ctx,
		r,
		existingHPA,
		hpa,
		log,
		capsule,
		status,
		func(t1, t2 *autoscalingv2.HorizontalPodAutoscaler) bool {
			return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
		},
	)
}

func shouldCreateHPA(capsule *v1alpha2.Capsule, scheme *runtime.Scheme) (bool, error) {
	_, res, err := createHPA(capsule, scheme)
	if err != nil {
		return false, err
	}
	return res, nil
}

func createHPA(
	capsule *v1alpha2.Capsule,
	scheme *runtime.Scheme,
) (*autoscalingv2.HorizontalPodAutoscaler, bool, error) {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      capsule.Name,
			Namespace: capsule.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       capsule.Name,
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		},
	}
	if err := controllerutil.SetControllerReference(capsule, hpa, scheme); err != nil {
		return nil, true, err
	}

	scale := capsule.Spec.Scale.Horizontal

	if scale.Instances.Min == 0 {
		// Cannot have autoscaler going to 0.
		// TODO We should have some good documentation/userfeedback if min-replicas is set to 0
		return hpa, false, nil
	}

	if scale.Instances.Max == nil {
		return hpa, false, nil
	}

	if scale.CPUTarget != nil && scale.CPUTarget.Utilization != nil {
		hpa.Spec.Metrics = append(hpa.Spec.Metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: v1.ResourceCPU,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: ptr.New(int32(*scale.CPUTarget.Utilization)),
				},
			},
		})
	}

	for _, customMetric := range scale.CustomMetrics {
		if customMetric.InstanceMetric != nil {
			instanceMetric := customMetric.InstanceMetric
			averageValue, err := resource.ParseQuantity(instanceMetric.AverageValue)
			if err != nil {
				return nil, false, err
			}
			metric := autoscalingv2.MetricSpec{
				Type: autoscalingv2.PodsMetricSourceType,
				Pods: &autoscalingv2.PodsMetricSource{
					Metric: autoscalingv2.MetricIdentifier{
						Name: instanceMetric.MetricName,
					},
					Target: autoscalingv2.MetricTarget{
						Type:         autoscalingv2.AverageValueMetricType,
						AverageValue: &averageValue,
					},
				},
			}
			if instanceMetric.MatchLabels != nil {
				metric.Pods.Metric.Selector.MatchLabels = instanceMetric.MatchLabels
			}
			hpa.Spec.Metrics = append(hpa.Spec.Metrics, metric)
		} else if customMetric.ObjectMetric != nil {
			object := customMetric.ObjectMetric
			metric := autoscalingv2.MetricSpec{
				Type: autoscalingv2.ObjectMetricSourceType,
				Object: &autoscalingv2.ObjectMetricSource{
					DescribedObject: object.DescribedObject,
					Metric: autoscalingv2.MetricIdentifier{
						Name: object.MetricName,
					},
				},
			}
			if object.AverageValue != "" {
				averageValue, err := resource.ParseQuantity(object.AverageValue)
				if err != nil {
					return nil, false, err
				}
				metric.Object.Target.Value = &averageValue
				metric.Object.Target.Type = autoscalingv2.AverageValueMetricType
			} else if object.Value != "" {
				value, err := resource.ParseQuantity(object.Value)
				if err != nil {
					return nil, false, err
				}
				metric.Object.Target.Value = &value
				metric.Object.Target.Type = autoscalingv2.ValueMetricType
			}
			if object.MatchLabels != nil {
				metric.Object.Metric.Selector.MatchLabels = object.MatchLabels
			}
			hpa.Spec.Metrics = append(hpa.Spec.Metrics, metric)
		}
	}

	if len(hpa.Spec.Metrics) == 0 {
		return hpa, false, nil
	}

	hpa.Spec.MinReplicas = ptr.New(int32(scale.Instances.Min))
	hpa.Spec.MaxReplicas = int32(*scale.Instances.Max)

	return hpa, true, nil
}

func (r *CapsuleReconciler) reconcileServiceAccount(
	ctx context.Context,
	_ ctrl.Request,
	log logr.Logger,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
) error {
	sa, err := createServiceAccount(capsule, r.Scheme)
	if err != nil {
		return err
	}

	existingSA := &v1.ServiceAccount{}
	if err = r.Get(ctx, client.ObjectKeyFromObject(sa), existingSA); err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("creating service account")
			if err := r.Create(ctx, sa); err != nil {
				return fmt.Errorf("could not create service account: %w", err)
			}
			return nil
		} else if err != nil {
			return fmt.Errorf("could not fetch service account: %w", err)
		}
	}

	return upsertIfNewer(ctx, r, existingSA, sa, log, capsule, status, func(t1, t2 *v1.ServiceAccount) bool {
		return equality.Semantic.DeepEqual(t1.Annotations, t2.Annotations)
	})
}

func createServiceAccount(capsule *v1alpha2.Capsule, scheme *runtime.Scheme) (*v1.ServiceAccount, error) {
	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      capsule.Name,
			Namespace: capsule.Namespace,
		},
	}
	if err := controllerutil.SetControllerReference(capsule, sa, scheme); err != nil {
		return nil, err
	}

	return sa, nil
}

func upsertIfNewer[T client.Object](
	ctx context.Context,
	r *CapsuleReconciler,
	currentObj T,
	newObj T,
	log logr.Logger,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
	equal func(t1 T, t2 T) bool,
) error {
	gvks, _, err := r.Scheme.ObjectKinds(currentObj)
	if err != nil {
		return fmt.Errorf("could not get object kinds for object: %w", err)
	}
	gvk := gvks[0]
	log = log.WithValues(
		"gvk", map[string]string{
			"kind":     gvk.Kind,
			"group":    gvk.Group,
			"version":  gvk.Version,
			"obj_name": newObj.GetName(),
		},
	)

	res := v1alpha2.OwnedResource{
		Ref: &v1.TypedLocalObjectReference{
			Kind: gvk.Kind,
			Name: newObj.GetName(),
		},
		State: "created",
	}
	defer func() {
		status.OwnedResources = append(status.OwnedResources, res)
	}()

	if !IsOwnedBy(capsule, newObj) {
		log.Info("Found existing resource not owned by capsule. Will not update it.")
		res.State = "failed"
		res.Message = "found existing resource not owned by capsule"
		return fmt.Errorf("found existing %s not owned by capsule", gvk.Kind)
	}

	materializedObj := newObj.DeepCopyObject().(T)

	// Dry run to fully materialize the new spec.
	materializedObj.SetResourceVersion(currentObj.GetResourceVersion())
	if err := r.Update(ctx, materializedObj, client.DryRunAll); err != nil {
		res.State = "failed"
		res.Message = err.Error()
		return fmt.Errorf("could not test update to %s: %w", gvk.Kind, err)
	}

	materializedObj.SetResourceVersion("")
	if !equal(materializedObj, currentObj) {
		log.Info("updating resource")
		if err := r.Update(ctx, newObj); err != nil {
			res.State = "failed"
			res.Message = err.Error()
			return fmt.Errorf("could not update %s: %w", gvk.Kind, err)
		}
		return nil
	}

	log.Info("resource is up-to-date")
	return nil
}

func (r *CapsuleReconciler) reconcilePrometheusServiceMonitor(
	ctx context.Context,
	_ ctrl.Request,
	log logr.Logger,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
) error {
	if r.Config.PrometheusServiceMonitor == nil || r.Config.PrometheusServiceMonitor.PortName == "" {
		return nil
	}

	serviceMonitor, err := r.createPrometheusServiceMonitor(capsule, r.Scheme)
	if err != nil {
		return err
	}

	existingServiceMonitor := &monitorv1.ServiceMonitor{}
	if err = r.Get(ctx, client.ObjectKeyFromObject(serviceMonitor), existingServiceMonitor); err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("creating prometheus service monitor")
			if err := r.Create(ctx, serviceMonitor); err != nil {
				return fmt.Errorf("could not create prometheus service monitor: %w", err)
			}
			return nil
		} else if err != nil {
			return fmt.Errorf("could not fetch prometheus service monitor: %w", err)
		}
	}

	return upsertIfNewer(
		ctx, r,
		existingServiceMonitor,
		serviceMonitor,
		log, capsule, status,
		func(t1, t2 *monitorv1.ServiceMonitor) bool {
			return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
		},
	)
}

func (r *CapsuleReconciler) createPrometheusServiceMonitor(
	capsule *v1alpha2.Capsule,
	scheme *runtime.Scheme,
) (*monitorv1.ServiceMonitor, error) {
	s := &monitorv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            capsule.Name,
			Namespace:       capsule.Namespace,
			ResourceVersion: "",
			Labels: map[string]string{
				LabelCapsule: capsule.Name,
			},
		},
		Spec: monitorv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					LabelCapsule: capsule.Name,
				},
			},
			Endpoints: []monitorv1.Endpoint{{
				Port: r.Config.PrometheusServiceMonitor.PortName,
				Path: r.Config.PrometheusServiceMonitor.Path,
			}},
		},
	}
	if err := controllerutil.SetControllerReference(capsule, s, scheme); err != nil {
		return nil, err
	}

	return s, nil
}

func (r *CapsuleReconciler) reconcileCronJobs(
	ctx context.Context,
	req ctrl.Request,
	log logr.Logger,
	capsule *v1alpha2.Capsule,
	status *v1alpha2.CapsuleStatus,
) error {
	configs, err := r.getConfigs(ctx, req, capsule, status)
	if err != nil {
		return err
	}

	checksums, err := r.configChecksums(capsule, configs)
	if err != nil {
		return err
	}

	jobs, err := r.createCronJobs(capsule, r.Scheme, configs, checksums)
	if err != nil {
		return err
	}

	existingJobs := &batchv1.CronJobList{}
	if err = r.List(ctx, existingJobs, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			LabelCapsule: capsule.Name,
		}),
	}); err != nil {
		return err
	}

	// Create/update jobs
	existingJobByName := map[string]batchv1.CronJob{}
	for _, j := range existingJobs.Items {
		existingJobByName[j.Name] = j
	}
	for _, job := range jobs {
		existingCronJob, ok := existingJobByName[job.Name]
		if !ok {
			log.Info("creating cron job", "name", job.Name)
			if err := r.Create(ctx, job); err != nil {
				return fmt.Errorf("could not create cron job %s: %w", job.Name, err)
			}
			continue
		}

		if err := upsertIfNewer(
			ctx, r,
			&existingCronJob, job,
			log, capsule, status,
			func(t1, t2 *batchv1.CronJob) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			},
		); err != nil {
			return err
		}
	}

	// Delete extraneous jobs
	nameOfJobs := map[string]struct{}{}
	for _, j := range jobs {
		nameOfJobs[j.Name] = struct{}{}
	}
	for _, j := range existingJobs.Items {
		if _, ok := nameOfJobs[j.Name]; !ok {
			if err := r.Delete(ctx, &j); err != nil {
				return fmt.Errorf("failed to delete cron job %s: %w", j.Name, err)
			}
		}
	}

	return nil
}

func (r *CapsuleReconciler) createCronJobs(
	capsule *v1alpha2.Capsule,
	scheme *runtime.Scheme,
	configs *configs,
	checksums *checksums,
) ([]*batchv1.CronJob, error) {
	var res []*batchv1.CronJob
	deployment, err := createDeployment(capsule, scheme, configs, checksums, nil)
	if err != nil {
		return nil, err
	}

	for _, job := range capsule.Spec.CronJobs {
		var template v1.PodTemplateSpec
		if job.Command != nil {
			template = *deployment.Spec.Template.DeepCopy()
			c := template.Spec.Containers[0]
			c.Command = []string{job.Command.Command}
			c.Args = job.Command.Args
			template.Spec.Containers[0] = c
			template.Spec.RestartPolicy = v1.RestartPolicyNever

		} else if job.URL != nil {
			template = createURLCronJobTemplate(capsule, job)
		} else {
			return nil, fmt.Errorf("neither Command nor URL was set on job %s", job.Name)
		}

		annotations := map[string]string{}
		maps.Copy(annotations, capsule.Annotations)

		j := &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", capsule.Name, job.Name),
				Namespace: capsule.Namespace,
				Labels: map[string]string{
					LabelCapsule: capsule.Name,
					LabelCron:    job.Name,
				},
				Annotations: annotations,
			},
			Spec: batchv1.CronJobSpec{
				Schedule: job.Schedule,
				JobTemplate: batchv1.JobTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: annotations,
						Labels: map[string]string{
							LabelCapsule: capsule.Name,
							LabelCron:    job.Name,
						},
					},
					Spec: batchv1.JobSpec{
						ActiveDeadlineSeconds: ptr.Convert[uint, int64](job.TimeoutSeconds),
						BackoffLimit:          ptr.Convert[uint, int32](job.MaxRetries),
						Template:              template,
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(capsule, j, scheme); err != nil {
			return nil, err
		}
		res = append(res, j)
	}

	return res, nil
}

func createURLCronJobTemplate(capsule *v1alpha2.Capsule, job v1alpha2.CronJob) v1.PodTemplateSpec {
	args := []string{"-G", "--fail-with-body"}
	for k, v := range job.URL.QueryParameters {
		args = append(args, "-d", fmt.Sprintf("%v=%v", url.QueryEscape(k), url.QueryEscape(v)))
	}
	urlString := fmt.Sprintf("http://%s:%v%s", capsule.Name, job.URL.Port, job.URL.Path)
	args = append(args, urlString)
	return v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name:    fmt.Sprintf("%s-%s", capsule.Name, job.Name),
				Image:   "quay.io/curl/curl:latest",
				Command: []string{"curl"},
				Args:    args,
			}},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
}
