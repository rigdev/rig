package controller

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/utils"
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeploymentStep struct{}

func NewDeploymentStep() *DeploymentStep {
	return &DeploymentStep{}
}

func (s *DeploymentStep) Apply(ctx context.Context, req Request) error {
	cfgs, err := s.getConfigs(ctx, req)
	if err != nil {
		return err
	}

	checksums, err := s.getConfigChecksums(req, *cfgs)
	if err != nil {
		return err
	}

	key := req.ObjectKey(_appsDeploymentGVK)
	current := GetCurrent[*appsv1.Deployment](req, key)

	deployment, err := s.createDeployment(current, req, cfgs, checksums)
	if err != nil {
		return err
	}

	req.Set(key, deployment)

	if ok, err := s.shouldCreateHPA(req); err != nil {
		return err
	} else if ok {
		hpa, _, err := s.createHPA(req)
		if err != nil {
			return err
		}

		req.Set(req.ObjectKey(_autoscalingvHorizontalPodAutoscalerGVK), hpa)
	}

	return nil
}

func (s *DeploymentStep) createDeployment(current *appsv1.Deployment, req Request, cfgs *configs, checksums checksums) (*appsv1.Deployment, error) {
	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount
	for _, f := range req.Capsule().Spec.Files {
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
	maps.Copy(podAnnotations, req.Capsule().Annotations)
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
	if req.Capsule().Spec.Env == nil || !req.Capsule().Spec.Env.DisableAutomatic {
		if _, ok := cfgs.configMaps[req.Capsule().GetName()]; ok {
			envFrom = append(envFrom, v1.EnvFromSource{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: req.Capsule().GetName()},
				},
			})
		}
		if _, ok := cfgs.secrets[req.Capsule().GetName()]; ok {
			envFrom = append(envFrom, v1.EnvFromSource{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: req.Capsule().GetName()},
				},
			})
		}
	}

	if req.Capsule().Spec.Env != nil {
		for _, e := range req.Capsule().Spec.Env.From {
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

	for _, name := range cfgs.sharedEnvConfigMaps {
		envFrom = append(envFrom, v1.EnvFromSource{
			ConfigMapRef: &v1.ConfigMapEnvSource{
				LocalObjectReference: v1.LocalObjectReference{Name: name},
			},
		})
	}
	for _, name := range cfgs.sharedEnvSecrets {
		envFrom = append(envFrom, v1.EnvFromSource{
			SecretRef: &v1.SecretEnvSource{
				LocalObjectReference: v1.LocalObjectReference{Name: name},
			},
		})
	}

	c := v1.Container{
		Name:         req.Capsule().Name,
		Image:        req.Capsule().Spec.Image,
		EnvFrom:      envFrom,
		VolumeMounts: volumeMounts,
		Resources:    makeResourceRequirements(req.Capsule()),
		Args:         req.Capsule().Spec.Args,
	}

	if req.Capsule().Spec.Command != "" {
		c.Command = []string{req.Capsule().Spec.Command}
	}

	replicas := ptr.New(int32(req.Capsule().Spec.Scale.Horizontal.Instances.Min))
	hasHPA, err := s.shouldCreateHPA(req)
	if err != nil {
		return nil, err
	}
	if hasHPA {
		if current != nil && current.Spec.Replicas != nil {
			replicas = ptr.New(*current.Spec.Replicas)
		}
	}
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					LabelCapsule: req.Capsule().Name,
				},
			},
			Replicas: replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: podAnnotations,
					Labels: map[string]string{
						LabelCapsule: req.Capsule().Name,
					},
				},
				Spec: v1.PodSpec{
					Containers:         []v1.Container{c},
					ServiceAccountName: req.Capsule().Name,
					Volumes:            volumes,
					NodeSelector:       req.Capsule().Spec.NodeSelector,
				},
			},
		},
	}

	return d, nil
}

func (s *DeploymentStep) getConfigChecksums(req Request, cfgs configs) (checksums, error) {
	sharedEnv, err := configSharedEnvChecksum(cfgs)
	if err != nil {
		return checksums{}, err
	}

	autoEnv, err := configAutoEnvChecksum(
		cfgs.configMaps[req.Capsule().GetName()],
		cfgs.secrets[req.Capsule().GetName()],
	)
	if err != nil {
		return checksums{}, err
	}

	env, err := configEnvChecksum(req, cfgs)
	if err != nil {
		return checksums{}, err
	}

	files, err := configFilesChecksum(req, cfgs)
	if err != nil {
		return checksums{}, err
	}

	return checksums{
		sharedEnv: sharedEnv,
		autoEnv:   autoEnv,
		env:       env,
		files:     files,
	}, nil
}

func configSharedEnvChecksum(cfgs configs) (string, error) {
	if !cfgs.hasSharedConfig() {
		return "", nil
	}

	h := sha256.New()

	configMaps := slices.Clone(cfgs.sharedEnvConfigMaps)
	slices.Sort(configMaps)
	secrets := slices.Clone(cfgs.sharedEnvSecrets)
	slices.Sort(secrets)

	for _, name := range configMaps {
		if err := hash.ConfigMap(h, cfgs.configMaps[name]); err != nil {
			return "", err
		}
	}
	for _, name := range secrets {
		if err := hash.Secret(h, cfgs.secrets[name]); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func configAutoEnvChecksum(
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

func configEnvChecksum(req Request, cfgs configs) (string, error) {
	if req.Capsule().Spec.Env == nil || len(req.Capsule().Spec.Env.From) == 0 {
		return "", nil
	}

	h := sha256.New()
	for _, e := range req.Capsule().Spec.Env.From {
		switch e.Kind {
		case "ConfigMap":
			if err := hash.ConfigMap(h, cfgs.configMaps[e.Name]); err != nil {
				return "", err
			}
		case "Secret":
			if err := hash.Secret(h, cfgs.secrets[e.Name]); err != nil {
				return "", err
			}
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func configFilesChecksum(req Request, cfgs configs) (string, error) {
	if len(req.Capsule().Spec.Files) == 0 {
		return "", nil
	}

	referencedKeysBySecretName := map[string]map[string]struct{}{}
	referencedKeysByConfigMapName := map[string]map[string]struct{}{}
	for _, f := range req.Capsule().Spec.Files {
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
			cfgs.secrets[name],
		); err != nil {
			return "", err
		}
	}
	for _, name := range configMapNames {
		if err := hash.ConfigMapKeys(
			h,
			maps.Keys(referencedKeysByConfigMapName[name]),
			cfgs.configMaps[name],
		); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (s *DeploymentStep) getConfigs(ctx context.Context, req Request) (*configs, error) {
	cfgs := &configs{
		configMaps: map[string]*v1.ConfigMap{},
		secrets:    map[string]*v1.Secret{},
	}

	// Get shared env
	var configMapList v1.ConfigMapList
	if err := req.Client().List(ctx, &configMapList, &client.ListOptions{
		Namespace: req.Capsule().Namespace,
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
	if err := req.Client().List(ctx, &secretList, &client.ListOptions{
		Namespace: req.Capsule().Namespace,
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

	env := req.Capsule().Spec.Env
	if env == nil {
		env = &v1alpha2.Env{}
	}

	// Get automatic env
	if !env.DisableAutomatic {
		if err := s.setUsedSource(ctx, req, cfgs, "ConfigMap", req.Capsule().Name, false); err != nil {
			return nil, err
		}

		if err := s.setUsedSource(ctx, req, cfgs, "Secret", req.Capsule().Name, false); err != nil {
			return nil, err
		}
	}

	// Get envs
	for _, e := range env.From {
		if err := s.setUsedSource(ctx, req, cfgs, e.Kind, e.Name, true); err != nil {
			return nil, err
		}
	}

	// Get files
	for _, f := range req.Capsule().Spec.Files {
		if err := s.setUsedSource(ctx, req, cfgs, f.Ref.Kind, f.Ref.Name, true); err != nil {
			return nil, err
		}
	}

	return cfgs, nil
}

func (s *DeploymentStep) setUsedSource(
	ctx context.Context,
	req Request,
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

		// r.status.UsedResources = append(r.status.UsedResources, ref)
	}()

	switch kind {
	case "ConfigMap":
		if _, ok := cfgs.configMaps[name]; ok {
			return nil
		}
		var cm v1.ConfigMap
		if err := req.Client().Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: req.Capsule().Namespace,
		}, &cm); err != nil {
			return fmt.Errorf("could not get referenced environment configmap: %w", err)
		}
		cfgs.configMaps[cm.Name] = &cm
	case "Secret":
		if _, ok := cfgs.secrets[name]; ok {
			return nil
		}
		var s v1.Secret
		if err := req.Client().Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: req.Capsule().Namespace,
		}, &s); err != nil {
			return fmt.Errorf("could not get referenced environment secret: %w", err)
		}
		cfgs.secrets[s.Name] = &s
	}

	return nil
}

func (s *DeploymentStep) shouldCreateHPA(req Request) (bool, error) {
	_, res, err := s.createHPA(req)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (s *DeploymentStep) createHPA(req Request) (*autoscalingv2.HorizontalPodAutoscaler, bool, error) {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       req.Capsule().Name,
				APIVersion: appsv1.SchemeGroupVersion.String(),
			},
		},
	}

	scale := req.Capsule().Spec.Scale.Horizontal

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
