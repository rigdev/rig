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

type reconcileRequest struct {
	scheme *runtime.Scheme
	config *configv1alpha1.OperatorConfig
	client client.Client

	req       ctrl.Request
	logger    logr.Logger
	capsule   v1alpha2.Capsule
	status    v1alpha2.CapsuleStatus
	checksums checksums
	configs   configs
}

type reconcileStepFunc func(
	ctx context.Context,
	r *reconcileRequest,
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
		reconcilerSetup,
		reconcileHorizontalPodAutoscaler,
		reconcileDeployment,
		reconcileService,
		reconcileCertificate,
		reconcileIngress,
		reconcileLoadBalancer,
		reconcileServiceAccount,
		reconcileCronJobs,
	}

	configEventHandler := handler.EnqueueRequestsFromMapFunc(findCapsulesForConfig(mgr))

	b := ctrl.NewControllerManagedBy(mgr)
	if hasServiceMonitor {
		r.reconcileSteps = append(r.reconcileSteps, reconcilePrometheusServiceMonitor)
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

	reconciler := &reconcileRequest{
		req:     req,
		logger:  log,
		capsule: v1alpha2.Capsule{},
		status: v1alpha2.CapsuleStatus{
			Deployment: &v1alpha2.DeploymentStatus{},
		},
		client: r.Client,
		scheme: r.Scheme,
		config: r.Config,
	}
	if err := r.Get(ctx, req.NamespacedName, &reconciler.capsule); err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("could not fetch Capsule: %w", err)
	}

	var stepErrs []error
	for _, sf := range r.reconcileSteps {
		if err := sf(ctx, reconciler); err != nil {
			stepErrs = append(stepErrs, err)
		}
	}

	if len(stepErrs) == 0 {
		reconciler.status.ObservedGeneration = reconciler.capsule.GetGeneration()
	} else {
		var errs []string
		for _, e := range stepErrs {
			errs = append(errs, e.Error())
		}
		reconciler.status.Errors = errs
	}

	reconciler.capsule.Status = &reconciler.status

	if err := r.Status().Update(ctx, &reconciler.capsule); err != nil {
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

func reconcilerSetup(ctx context.Context, r *reconcileRequest) error {
	if err := r.setConfigs(ctx); err != nil {
		return err
	}
	if err := r.setConfigChecksums(); err != nil {
		return err
	}

	return nil
}

func (r *reconcileRequest) setConfigChecksums() error {
	sharedEnv, err := r.configSharedEnvChecksum()
	if err != nil {
		return err
	}

	autoEnv, err := r.configAutoEnvChecksum(
		r.configs.configMaps[r.capsule.GetName()],
		r.configs.secrets[r.capsule.GetName()],
	)
	if err != nil {
		return err
	}

	env, err := r.configEnvChecksum()
	if err != nil {
		return err
	}

	files, err := r.configFilesChecksum()
	if err != nil {
		return err
	}

	r.checksums = checksums{
		sharedEnv: sharedEnv,
		autoEnv:   autoEnv,
		env:       env,
		files:     files,
	}

	return nil
}

func (r *reconcileRequest) configSharedEnvChecksum() (string, error) {
	if !r.configs.hasSharedConfig() {
		return "", nil
	}

	h := sha256.New()

	configMaps := slices.Clone(r.configs.sharedEnvConfigMaps)
	slices.Sort(configMaps)
	secrets := slices.Clone(r.configs.sharedEnvSecrets)
	slices.Sort(secrets)

	for _, name := range configMaps {
		if err := hash.ConfigMap(h, r.configs.configMaps[name]); err != nil {
			return "", err
		}
	}
	for _, name := range secrets {
		if err := hash.Secret(h, r.configs.secrets[name]); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *reconcileRequest) configAutoEnvChecksum(
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

func (r *reconcileRequest) configEnvChecksum() (string, error) {
	if r.capsule.Spec.Env == nil || len(r.capsule.Spec.Env.From) == 0 {
		return "", nil
	}

	h := sha256.New()
	for _, e := range r.capsule.Spec.Env.From {
		switch e.Kind {
		case "ConfigMap":
			if err := hash.ConfigMap(h, r.configs.configMaps[e.Name]); err != nil {
				return "", err
			}
		case "Secret":
			if err := hash.Secret(h, r.configs.secrets[e.Name]); err != nil {
				return "", err
			}
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *reconcileRequest) configFilesChecksum() (string, error) {
	if len(r.capsule.Spec.Files) == 0 {
		return "", nil
	}

	referencedKeysBySecretName := map[string]map[string]struct{}{}
	referencedKeysByConfigMapName := map[string]map[string]struct{}{}
	for _, f := range r.capsule.Spec.Files {
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
			r.configs.secrets[name],
		); err != nil {
			return "", err
		}
	}
	for _, name := range configMapNames {
		if err := hash.ConfigMapKeys(
			h,
			maps.Keys(referencedKeysByConfigMapName[name]),
			r.configs.configMaps[name],
		); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (r *reconcileRequest) setConfigs(ctx context.Context) error {
	r.configs = configs{
		configMaps: map[string]*v1.ConfigMap{},
		secrets:    map[string]*v1.Secret{},
	}

	// Get shared env
	var configMapList v1.ConfigMapList
	if err := r.client.List(ctx, &configMapList, &client.ListOptions{
		Namespace: r.req.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			LabelSharedConfig: "true",
		}),
	}); err != nil {
		return fmt.Errorf("could not list shared env configmaps: %w", err)
	}
	r.configs.sharedEnvConfigMaps = make([]string, len(configMapList.Items))
	for i, cm := range configMapList.Items {
		r.configs.sharedEnvConfigMaps[i] = cm.GetName()
		r.configs.configMaps[cm.Name] = &cm
	}
	var secretList v1.SecretList
	if err := r.client.List(ctx, &secretList, &client.ListOptions{
		Namespace: r.req.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			LabelSharedConfig: "true",
		}),
	}); err != nil {
		return fmt.Errorf("could not list shared env secrets: %w", err)
	}
	r.configs.sharedEnvSecrets = make([]string, len(secretList.Items))
	for i, s := range secretList.Items {
		r.configs.sharedEnvSecrets[i] = s.GetName()
		r.configs.secrets[s.Name] = &s
	}

	env := r.capsule.Spec.Env
	if env == nil {
		env = &v1alpha2.Env{}
	}

	// Get automatic env
	if !env.DisableAutomatic {
		if err := r.setUsedSource(ctx, "ConfigMap", r.req.NamespacedName.Name, false); err != nil {
			return err
		}

		if err := r.setUsedSource(ctx, "Secret", r.req.NamespacedName.Name, false); err != nil {
			return err
		}
	}

	// Get envs
	for _, e := range env.From {
		if err := r.setUsedSource(ctx, e.Kind, e.Name, true); err != nil {
			return err
		}
	}

	// Get files
	for _, f := range r.capsule.Spec.Files {
		if err := r.setUsedSource(ctx, f.Ref.Kind, f.Ref.Name, true); err != nil {
			return err
		}
	}

	return nil
}

func (r *reconcileRequest) setUsedSource(
	ctx context.Context,
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

		r.status.UsedResources = append(r.status.UsedResources, ref)
	}()

	switch kind {
	case "ConfigMap":
		if _, ok := r.configs.configMaps[name]; ok {
			return nil
		}
		var cm v1.ConfigMap
		if err := r.client.Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: r.capsule.Namespace,
		}, &cm); err != nil {
			return fmt.Errorf("could not get referenced environment configmap: %w", err)
		}
		r.configs.configMaps[cm.Name] = &cm
	case "Secret":
		if _, ok := r.configs.secrets[name]; ok {
			return nil
		}
		var s v1.Secret
		if err := r.client.Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: r.capsule.Namespace,
		}, &s); err != nil {
			return fmt.Errorf("could not get referenced environment secret: %w", err)
		}
		r.configs.secrets[s.Name] = &s
	}

	return nil
}

func reconcileDeployment(ctx context.Context, r *reconcileRequest) error {
	existingDeploy := &appsv1.Deployment{}
	hasExistingDeployment := true
	if err := r.client.Get(ctx, r.req.NamespacedName, existingDeploy); err != nil {
		if kerrors.IsNotFound(err) {
			hasExistingDeployment = false
		} else {
			r.status.Deployment.State = "failed"
			r.status.Deployment.Message = err.Error()
			return fmt.Errorf("could not fetch deployment: %w", err)
		}
	}

	deploy, err := r.createDeployment(existingDeploy)
	if err != nil {
		return err
	}

	if !hasExistingDeployment {
		r.logger.Info("creating deployment")
		if err := r.client.Create(ctx, deploy); err != nil {
			r.status.Deployment.State = "failed"
			r.status.Deployment.Message = err.Error()
			return fmt.Errorf("could not create deployment: %w", err)
		}
		existingDeploy = deploy
	}

	if err != nil {
		r.status.Deployment.State = "failed"
		r.status.Deployment.Message = err.Error()
		return err
	}

	// Edge case, this property is not carried over by k8s.
	delete(existingDeploy.Spec.Template.Annotations, "kubectl.kubernetes.io/restartedAt")

	err = upsertIfNewer(ctx, r, existingDeploy, deploy, func(t1, t2 *appsv1.Deployment) bool {
		return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
	})
	if err != nil {
		r.status.Deployment.State = "failed"
		r.status.Deployment.Message = err.Error()
	}
	return err
}

func (r *reconcileRequest) createDeployment(existingDeployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	var ports []v1.ContainerPort
	for _, i := range r.capsule.Spec.Interfaces {
		ports = append(ports, v1.ContainerPort{
			Name:          i.Name,
			ContainerPort: i.Port,
		})
	}

	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount
	for _, f := range r.capsule.Spec.Files {
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
	maps.Copy(podAnnotations, r.capsule.Annotations)
	if r.checksums.files != "" {
		podAnnotations[AnnotationChecksumFiles] = r.checksums.files
	}
	if r.checksums.autoEnv != "" {
		podAnnotations[AnnotationChecksumAutoEnv] = r.checksums.autoEnv
	}
	if r.checksums.env != "" {
		podAnnotations[AnnotationChecksumEnv] = r.checksums.env
	}
	if r.checksums.sharedEnv != "" {
		podAnnotations[AnnotationChecksumSharedEnv] = r.checksums.sharedEnv
	}

	var envFrom []v1.EnvFromSource
	if r.capsule.Spec.Env == nil || !r.capsule.Spec.Env.DisableAutomatic {
		if _, ok := r.configs.configMaps[r.capsule.GetName()]; ok {
			envFrom = append(envFrom, v1.EnvFromSource{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: r.capsule.GetName()},
				},
			})
		}
		if _, ok := r.configs.secrets[r.capsule.GetName()]; ok {
			envFrom = append(envFrom, v1.EnvFromSource{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: r.capsule.GetName()},
				},
			})
		}
	}

	if r.capsule.Spec.Env != nil {
		for _, e := range r.capsule.Spec.Env.From {
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

	for _, name := range r.configs.sharedEnvConfigMaps {
		envFrom = append(envFrom, v1.EnvFromSource{
			ConfigMapRef: &v1.ConfigMapEnvSource{
				LocalObjectReference: v1.LocalObjectReference{Name: name},
			},
		})
	}
	for _, name := range r.configs.sharedEnvSecrets {
		envFrom = append(envFrom, v1.EnvFromSource{
			SecretRef: &v1.SecretEnvSource{
				LocalObjectReference: v1.LocalObjectReference{Name: name},
			},
		})
	}

	c := v1.Container{
		Name:         r.capsule.Name,
		Image:        r.capsule.Spec.Image,
		EnvFrom:      envFrom,
		VolumeMounts: volumeMounts,
		Ports:        ports,
		Resources:    makeResourceRequirements(&r.capsule),
		Args:         r.capsule.Spec.Args,
	}

	if r.capsule.Spec.Command != "" {
		c.Command = []string{r.capsule.Spec.Command}
	}

	for _, i := range r.capsule.Spec.Interfaces {
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

	replicas := ptr.New(int32(r.capsule.Spec.Scale.Horizontal.Instances.Min))
	hasHPA, err := r.shouldCreateHPA()
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
			Name:      r.capsule.Name,
			Namespace: r.capsule.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					LabelCapsule: r.capsule.Name,
				},
			},
			Replicas: replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: podAnnotations,
					Labels: map[string]string{
						LabelCapsule: r.capsule.Name,
					},
				},
				Spec: v1.PodSpec{
					Containers:         []v1.Container{c},
					ServiceAccountName: r.capsule.Name,
					Volumes:            volumes,
					NodeSelector:       r.capsule.Spec.NodeSelector,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(&r.capsule, d, r.scheme); err != nil {
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

func reconcileService(ctx context.Context, r *reconcileRequest) error {
	service, err := r.createService()
	if err != nil {
		return err
	}

	existingService := &v1.Service{}
	if err := r.client.Get(ctx, r.req.NamespacedName, existingService); err != nil {
		if kerrors.IsNotFound(err) {
			if len(r.capsule.Spec.Interfaces) == 0 {
				return nil
			}

			r.logger.Info("creating service")
			if err := r.client.Create(ctx, service); err != nil {
				return fmt.Errorf("could not create service: %w", err)
			}
			existingService = service
		} else {
			return fmt.Errorf("could not fetch service: %w", err)
		}
	}

	if !IsOwnedBy(&r.capsule, existingService) {
		if len(r.capsule.Spec.Interfaces) == 0 {
			r.logger.Info("Found existing service not owned by capsule. Will not delete it.")
		} else {
			r.logger.Info("Found existing service not owned by capsule. Will not update it.")
		}
	} else {
		if len(r.capsule.Spec.Interfaces) == 0 {
			r.logger.Info("deleting service")
			if err := r.client.Delete(ctx, existingService); err != nil {
				return fmt.Errorf("could not delete service: %w", err)
			}
		} else {
			return upsertIfNewer(ctx, r, existingService, service, func(t1, t2 *v1.Service) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		}
	}

	return nil
}

func (r *reconcileRequest) createService() (*v1.Service, error) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.capsule.Name,
			Namespace: r.capsule.Namespace,
			Labels: map[string]string{
				LabelCapsule: r.capsule.Name,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				LabelCapsule: r.capsule.Name,
			},
			Type: r.config.Service.Type,
		},
	}

	for _, inf := range r.capsule.Spec.Interfaces {
		svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
			Name:       inf.Name,
			Port:       inf.Port,
			TargetPort: intstr.FromString(inf.Name),
		})
	}

	if err := controllerutil.SetControllerReference(&r.capsule, svc, r.scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on service: %w", err)
	}

	return svc, nil
}

func reconcileCertificate(ctx context.Context, r *reconcileRequest) error {
	crt, err := r.createCertificate()
	if err != nil {
		return err
	}

	existingCrt := &cmv1.Certificate{}
	if err := r.client.Get(ctx, r.req.NamespacedName, existingCrt); err != nil {
		if kerrors.IsNotFound(err) {
			if !r.capsuleHasIngress() {
				return nil
			}
			if !r.ingressIsSupported() {
				r.logger.V(1).Info("not creating certificate as ingress is not supported: cert-manager config missing")
				return nil
			}
			if !r.shouldCreateCertificateRessource() {
				r.logger.V(1).Info("not creating certificate as operator is configured to use ingress annotations")
				return nil
			}

			r.logger.Info("creating certificate")
			if err := r.client.Create(ctx, crt); err != nil {
				return fmt.Errorf("could not create certificate: %w", err)
			}
			existingCrt = crt
		} else {
			return fmt.Errorf("could not fetch certificate: %w", err)
		}
	}

	if !IsOwnedBy(&r.capsule, existingCrt) {
		if r.capsuleHasIngress() {
			r.logger.Info("Found existing certificate not owned by capsule. Will not update it.")
			return errors.New("found existing certificate not owned by capsule")
		}
		r.logger.Info("Found existing certificate not owned by capsule. Will not delete it.")
	} else {
		if r.ingressIsSupported() && r.shouldCreateCertificateRessource() && r.capsuleHasIngress() {
			return upsertIfNewer(ctx, r, existingCrt, crt, func(t1, t2 *cmv1.Certificate) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		}
		if !r.ingressIsSupported() {
			r.logger.V(1).Info("deleting certificate as ingress is not supported: cert-manager config missing")
		} else if !r.shouldCreateCertificateRessource() {
			r.logger.V(1).Info("deleting certificate becausee operator is configured to use ingress annotations")
		} else {
			r.logger.Info("deleting certificate")
		}
		if err := r.client.Delete(ctx, existingCrt); err != nil {
			return fmt.Errorf("could not delete certificate: %w", err)
		}
	}

	return nil
}

func (r *reconcileRequest) shouldCreateCertificateRessource() bool {
	return r.config.Certmanager != nil &&
		r.config.Certmanager.CreateCertificateResources &&
		!r.config.Ingress.DisableTLS
}

func (r *reconcileRequest) createCertificate() (*cmv1.Certificate, error) {
	crt := &cmv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.capsule.Name,
			Namespace: r.capsule.Namespace,
		},
		Spec: cmv1.CertificateSpec{
			SecretName: fmt.Sprintf("%s-tls", r.capsule.Name),
		},
	}

	if r.config.Certmanager != nil {
		crt.Spec.IssuerRef = cmmetav1.ObjectReference{
			Kind: cmv1.ClusterIssuerKind,
			Name: r.config.Certmanager.ClusterIssuer,
		}
	}

	for _, inf := range r.capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			crt.Spec.DNSNames = append(crt.Spec.DNSNames, inf.Public.Ingress.Host)
		}
	}

	if err := controllerutil.SetControllerReference(&r.capsule, crt, r.scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on certificate: %w", err)
	}

	return crt, nil
}

func (r *reconcileRequest) ingressIsSupported() bool {
	return r.config.Ingress.DisableTLS || (r.config.Certmanager != nil && r.config.Certmanager.ClusterIssuer != "")
}

func reconcileIngress(ctx context.Context, r *reconcileRequest) error {
	ing, err := r.createIngress()
	if err != nil {
		return err
	}

	existingIng := &netv1.Ingress{}
	if err := r.client.Get(ctx, r.req.NamespacedName, existingIng); err != nil {
		if kerrors.IsNotFound(err) {
			if !r.capsuleHasIngress() {
				return nil
			}
			if !r.ingressIsSupported() {
				r.logger.V(1).Info("ingress not supported: cert-manager config missing")
				return nil
			}

			r.logger.Info("creating ingress")
			if err := r.client.Create(ctx, ing); err != nil {
				return fmt.Errorf("could not create ingress: %w", err)
			}
			existingIng = ing
		} else {
			return fmt.Errorf("could not fetch ingress: %w", err)
		}
	}

	if !IsOwnedBy(&r.capsule, existingIng) {
		if r.capsuleHasIngress() {
			r.logger.Info("Found existing ingress not owned by capsule. Will not update it.")
			return errors.New("found existing ingress not owned by capsule")
		}
		r.logger.Info("Found existing ingress not owned by capsule. Will not delete it.")
	} else {
		if r.ingressIsSupported() && r.capsuleHasIngress() {
			return upsertIfNewer(ctx, r, existingIng, ing, func(t1, t2 *netv1.Ingress) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		}
		if !r.ingressIsSupported() {
			r.logger.V(1).Info("ingress not supported: cert-manager config missing")
		}
		r.logger.Info("deleting ingress")
		if err := r.client.Delete(ctx, existingIng); err != nil {
			return fmt.Errorf("could not delete ingress: %w", err)
		}
	}

	return nil
}

func (r *reconcileRequest) capsuleHasIngress() bool {
	for _, inf := range r.capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.Ingress != nil {
			return true
		}
	}
	return false
}

func (r *reconcileRequest) createIngress() (*netv1.Ingress, error) {
	ing := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.capsule.Name,
			Namespace:   r.capsule.Namespace,
			Annotations: r.config.Ingress.Annotations,
		},
	}

	if r.config.Ingress.ClassName != "" {
		ing.Spec.IngressClassName = ptr.New(r.config.Ingress.ClassName)
	}

	if r.ingressIsSupported() && !r.config.Ingress.DisableTLS && !r.shouldCreateCertificateRessource() {
		ing.Annotations["cert-manager.io/cluster-issuer"] = r.config.Certmanager.ClusterIssuer
	}

	for _, inf := range r.capsule.Spec.Interfaces {
		if inf.Public == nil || inf.Public.Ingress == nil {
			continue
		}

		ing.Spec.Rules = append(ing.Spec.Rules, netv1.IngressRule{
			Host: inf.Public.Ingress.Host,
			IngressRuleValue: netv1.IngressRuleValue{
				HTTP: &netv1.HTTPIngressRuleValue{},
			},
		})

		if len(inf.Public.Ingress.Paths) == 0 {
			ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths = []netv1.HTTPIngressPath{
				{
					PathType: ptr.New(r.config.Ingress.PathType),
					Path:     "/",
					Backend: netv1.IngressBackend{
						Service: &netv1.IngressServiceBackend{
							Name: r.capsule.Name,
							Port: netv1.ServiceBackendPort{
								Name: inf.Name,
							},
						},
					},
				},
			}
		} else {
			for _, path := range inf.Public.Ingress.Paths {
				ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths = append(
					ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths,
					netv1.HTTPIngressPath{
						PathType: ptr.New(r.config.Ingress.PathType),
						Path:     path,
						Backend: netv1.IngressBackend{
							Service: &netv1.IngressServiceBackend{
								Name: r.capsule.Name,
								Port: netv1.ServiceBackendPort{
									Name: inf.Name,
								},
							},
						},
					},
				)
			}
		}

		if !r.config.Ingress.DisableTLS {
			if len(ing.Spec.TLS) == 0 {
				ing.Spec.TLS = []netv1.IngressTLS{{
					SecretName: fmt.Sprintf("%s-tls", r.capsule.Name),
				}}
			}
			ing.Spec.TLS[0].Hosts = append(ing.Spec.TLS[0].Hosts, inf.Public.Ingress.Host)
		}
	}

	if err := controllerutil.SetControllerReference(&r.capsule, ing, r.scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on ingress: %w", err)
	}

	return ing, nil
}

func reconcileLoadBalancer(ctx context.Context, r *reconcileRequest) error {
	svc, err := r.createLoadBalancer()
	if err != nil {
		return err
	}

	nsName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-lb", r.req.NamespacedName.Name),
		Namespace: r.req.NamespacedName.Namespace,
	}
	existingSvc := &v1.Service{}
	if err := r.client.Get(ctx, nsName, existingSvc); err != nil {
		if kerrors.IsNotFound(err) {
			if !r.capsuleHasLoadBalancer() {
				return nil
			}

			r.logger.Info("creating loadbalancer service")
			if err := r.client.Create(ctx, svc); err != nil {
				return fmt.Errorf("could not create loadbalancer: %w", err)
			}
			existingSvc = svc
		} else {
			return fmt.Errorf("could not fetch loadbalancer: %w", err)
		}
	}

	if !IsOwnedBy(&r.capsule, existingSvc) {
		if r.capsuleHasLoadBalancer() {
			r.logger.Info("Found existing loadbalancer service not owned by capsule. Will not update it.")
			return errors.New("found existing loadbalancer service not owned by capsule")
		}
		r.logger.Info("Found existing loadbalancer service not owned by capsule. Will not delete it.")
	} else {
		if r.capsuleHasLoadBalancer() {
			return upsertIfNewer(ctx, r, existingSvc, svc, func(t1, t2 *v1.Service) bool {
				return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
			})
		}
		r.logger.Info("deleting loadbalancer service")
		if err := r.client.Delete(ctx, existingSvc); err != nil {
			return fmt.Errorf("could not delete loadbalancer service: %w", err)
		}
	}

	return nil
}

func (r *reconcileRequest) capsuleHasLoadBalancer() bool {
	for _, inf := range r.capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.LoadBalancer != nil {
			return true
		}
	}
	return false
}

func (r *reconcileRequest) createLoadBalancer() (*v1.Service, error) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-lb", r.capsule.Name),
			Namespace: r.capsule.Namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				LabelCapsule: r.capsule.Name,
			},
		},
	}

	for _, inf := range r.capsule.Spec.Interfaces {
		if inf.Public != nil && inf.Public.LoadBalancer != nil {
			svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
				Name:       inf.Name,
				Port:       inf.Public.LoadBalancer.Port,
				TargetPort: intstr.FromString(inf.Name),
			})
		}
	}

	if err := controllerutil.SetControllerReference(&r.capsule, svc, r.scheme); err != nil {
		return nil, fmt.Errorf("could not set owner reference on loadbalancer service: %w", err)
	}

	return svc, nil
}

func reconcileHorizontalPodAutoscaler(ctx context.Context, r *reconcileRequest) error {
	hpa, shouldHaveHPA, err := r.createHPA()
	if err != nil {
		return err
	}
	existingHPA := &autoscalingv2.HorizontalPodAutoscaler{}
	if err = r.client.Get(ctx, client.ObjectKeyFromObject(hpa), existingHPA); err != nil {
		if kerrors.IsNotFound(err) {
			if shouldHaveHPA {
				r.logger.Info("creating horizontal pod autoscaler")
				if err := r.client.Create(ctx, hpa); err != nil {
					return fmt.Errorf("could not create horizontal pod autoscaler: %w", err)
				}
			}
			return nil
		} else if err != nil {
			return fmt.Errorf("could not fetch horizontal pod autoscaler: %w", err)
		}
	}

	if !shouldHaveHPA {
		if err := r.client.Delete(ctx, existingHPA); err != nil {
			return err
		}
	}

	return upsertIfNewer(
		ctx,
		r,
		existingHPA,
		hpa,
		func(t1, t2 *autoscalingv2.HorizontalPodAutoscaler) bool {
			return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
		},
	)
}

func (r *reconcileRequest) shouldCreateHPA() (bool, error) {
	_, res, err := r.createHPA()
	if err != nil {
		return false, err
	}
	return res, nil
}

func (r *reconcileRequest) createHPA() (*autoscalingv2.HorizontalPodAutoscaler, bool, error) {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.capsule.Name,
			Namespace: r.capsule.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       r.capsule.Name,
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		},
	}
	if err := controllerutil.SetControllerReference(&r.capsule, hpa, r.scheme); err != nil {
		return nil, true, err
	}

	scale := r.capsule.Spec.Scale.Horizontal

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

func reconcileServiceAccount(ctx context.Context, r *reconcileRequest) error {
	sa, err := r.createServiceAccount()
	if err != nil {
		return err
	}

	existingSA := &v1.ServiceAccount{}
	if err = r.client.Get(ctx, client.ObjectKeyFromObject(sa), existingSA); err != nil {
		if kerrors.IsNotFound(err) {
			r.logger.Info("creating service account")
			if err := r.client.Create(ctx, sa); err != nil {
				return fmt.Errorf("could not create service account: %w", err)
			}
			return nil
		} else if err != nil {
			return fmt.Errorf("could not fetch service account: %w", err)
		}
	}

	return upsertIfNewer(ctx, r, existingSA, sa, func(t1, t2 *v1.ServiceAccount) bool {
		return equality.Semantic.DeepEqual(t1.Annotations, t2.Annotations)
	})
}

func (r *reconcileRequest) createServiceAccount() (*v1.ServiceAccount, error) {
	sa := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.capsule.Name,
			Namespace: r.capsule.Namespace,
		},
	}
	if err := controllerutil.SetControllerReference(&r.capsule, sa, r.scheme); err != nil {
		return nil, err
	}

	return sa, nil
}

func upsertIfNewer[T client.Object](
	ctx context.Context,
	r *reconcileRequest,
	currentObj T,
	newObj T,
	equal func(t1 T, t2 T) bool,
) error {
	gvks, _, err := r.scheme.ObjectKinds(currentObj)
	if err != nil {
		return fmt.Errorf("could not get object kinds for object: %w", err)
	}
	gvk := gvks[0]
	log := r.logger.WithValues(
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
		r.status.OwnedResources = append(r.status.OwnedResources, res)
	}()

	if !IsOwnedBy(&r.capsule, newObj) {
		log.Info("Found existing resource not owned by capsule. Will not update it.")
		res.State = "failed"
		res.Message = "found existing resource not owned by capsule"
		return fmt.Errorf("found existing %s not owned by capsule", gvk.Kind)
	}

	materializedObj := newObj.DeepCopyObject().(T)

	// Dry run to fully materialize the new spec.
	materializedObj.SetResourceVersion(currentObj.GetResourceVersion())
	if err := r.client.Update(ctx, materializedObj, client.DryRunAll); err != nil {
		res.State = "failed"
		res.Message = err.Error()
		return fmt.Errorf("could not test update to %s: %w", gvk.Kind, err)
	}

	materializedObj.SetResourceVersion("")
	if !equal(materializedObj, currentObj) {
		log.Info("updating resource")
		if err := r.client.Update(ctx, newObj); err != nil {
			res.State = "failed"
			res.Message = err.Error()
			return fmt.Errorf("could not update %s: %w", gvk.Kind, err)
		}
		return nil
	}

	log.Info("resource is up-to-date")
	return nil
}

func reconcilePrometheusServiceMonitor(ctx context.Context, r *reconcileRequest) error {
	if r.config.PrometheusServiceMonitor == nil || r.config.PrometheusServiceMonitor.PortName == "" {
		return nil
	}

	serviceMonitor, err := r.createPrometheusServiceMonitor()
	if err != nil {
		return err
	}

	existingServiceMonitor := &monitorv1.ServiceMonitor{}
	if err = r.client.Get(ctx, client.ObjectKeyFromObject(serviceMonitor), existingServiceMonitor); err != nil {
		if kerrors.IsNotFound(err) {
			r.logger.Info("creating prometheus service monitor")
			if err := r.client.Create(ctx, serviceMonitor); err != nil {
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
		func(t1, t2 *monitorv1.ServiceMonitor) bool {
			return equality.Semantic.DeepEqual(t1.Spec, t2.Spec)
		},
	)
}

func (r *reconcileRequest) createPrometheusServiceMonitor() (*monitorv1.ServiceMonitor, error) {
	s := &monitorv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            r.capsule.Name,
			Namespace:       r.capsule.Namespace,
			ResourceVersion: "",
			Labels: map[string]string{
				LabelCapsule: r.capsule.Name,
			},
		},
		Spec: monitorv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					LabelCapsule: r.capsule.Name,
				},
			},
			Endpoints: []monitorv1.Endpoint{{
				Port: r.config.PrometheusServiceMonitor.PortName,
				Path: r.config.PrometheusServiceMonitor.Path,
			}},
		},
	}
	if err := controllerutil.SetControllerReference(&r.capsule, s, r.scheme); err != nil {
		return nil, err
	}

	return s, nil
}

func reconcileCronJobs(ctx context.Context, r *reconcileRequest) error {
	jobs, err := r.createCronJobs()
	if err != nil {
		return err
	}

	existingJobs := &batchv1.CronJobList{}
	if err = r.client.List(ctx, existingJobs, &client.ListOptions{
		Namespace: r.req.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			LabelCapsule: r.capsule.Name,
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
			r.logger.Info("creating cron job", "name", job.Name)
			if err := r.client.Create(ctx, job); err != nil {
				return fmt.Errorf("could not create cron job %s: %w", job.Name, err)
			}
			continue
		}

		if err := upsertIfNewer(
			ctx, r,
			&existingCronJob, job,
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
			if err := r.client.Delete(ctx, &j); err != nil {
				return fmt.Errorf("failed to delete cron job %s: %w", j.Name, err)
			}
		}
	}

	return nil
}

func (r *reconcileRequest) createCronJobs() ([]*batchv1.CronJob, error) {
	var res []*batchv1.CronJob
	deployment, err := r.createDeployment(nil)
	if err != nil {
		return nil, err
	}

	for _, job := range r.capsule.Spec.CronJobs {
		var template v1.PodTemplateSpec
		if job.Command != nil {
			template = *deployment.Spec.Template.DeepCopy()
			c := template.Spec.Containers[0]
			c.Command = []string{job.Command.Command}
			c.Args = job.Command.Args
			template.Spec.Containers[0] = c
			template.Spec.RestartPolicy = v1.RestartPolicyNever

		} else if job.URL != nil {
			args := []string{"-G", "--fail-with-body"}
			for k, v := range job.URL.QueryParameters {
				args = append(args, "-d", fmt.Sprintf("%v=%v", url.QueryEscape(k), url.QueryEscape(v)))
			}
			urlString := fmt.Sprintf("http://%s:%v%s", r.capsule.Name, job.URL.Port, job.URL.Path)
			args = append(args, urlString)
			template = v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:    fmt.Sprintf("%s-%s", r.capsule.Name, job.Name),
						Image:   "quay.io/curl/curl:latest",
						Command: []string{"curl"},
						Args:    args,
					}},
					RestartPolicy: v1.RestartPolicyNever,
				},
			}
		} else {
			return nil, fmt.Errorf("neither Command nor URL was set on job %s", job.Name)
		}

		annotations := map[string]string{}
		maps.Copy(annotations, r.capsule.Annotations)

		j := &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", r.capsule.Name, job.Name),
				Namespace: r.capsule.Namespace,
				Labels: map[string]string{
					LabelCapsule: r.capsule.Name,
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
							LabelCapsule: r.capsule.Name,
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
		if err := controllerutil.SetControllerReference(&r.capsule, j, r.scheme); err != nil {
			return nil, err
		}
		res = append(res, j)
	}

	return res, nil
}
