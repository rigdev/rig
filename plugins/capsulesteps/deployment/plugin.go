// +groupName=plugins.rig.dev -- Only used for config doc generation
//
//nolint:revive
package deployment

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/pipeline"
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
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Name = "rigdev.deployment"

	AnnotationRecreateStrategy = "rig.dev/recreate-strategy"
	AnnotationEmptyDirs        = "rig.dev/empty-dirs"
)

// Configuration for the deployment plugin
// +kubebuilder:object:root=true
type Config struct{}

type Plugin struct {
	configBytes []byte
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	// We do not have any configuration for this step?

	// var config Config
	var err error
	if len(p.configBytes) > 0 {
		_, err = plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep[Config])
		if err != nil {
			return err
		}
	}

	cfgs, err := p.getConfigs(ctx, req)
	if err != nil {
		return err
	}

	checksums, err := p.getConfigChecksums(req, *cfgs)
	if err != nil {
		return err
	}

	current := &appsv1.Deployment{}
	if err := req.GetExistingInto(current); errors.IsNotFound(err) {
		current = nil
	} else if err != nil {
		return err
	}

	deployment, err := p.createDeployment(current, req, cfgs, checksums)
	if err != nil {
		return err
	}

	if len(req.Capsule().Spec.Interfaces) > 0 {
		if err := p.handleInterfaces(req, deployment); err != nil {
			return err
		}

		if err := req.Set(p.createService(req)); err != nil {
			return err
		}
	}

	if err := req.Set(deployment); err != nil {
		return err
	}

	if ok, err := p.shouldCreateHPA(req); err != nil {
		return err
	} else if ok {
		hpa, _, err := p.createHPA(req)
		if err != nil {
			return err
		}

		if err := req.Set(hpa); err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) createDeployment(
	current *appsv1.Deployment,
	req pipeline.CapsuleRequest,
	cfgs *configs,
	checksums checksums,
) (*appsv1.Deployment, error) {
	volumes, volumeMounts := pipeline.FilesToVolumes(req.Capsule().Spec.Files)

	podAnnotations := pipeline.CreatePodAnnotations(req)
	if checksums.files != "" {
		podAnnotations[pipeline.AnnotationChecksumFiles] = checksums.files
	}
	if checksums.autoEnv != "" {
		podAnnotations[pipeline.AnnotationChecksumAutoEnv] = checksums.autoEnv
	}
	if checksums.env != "" {
		podAnnotations[pipeline.AnnotationChecksumEnv] = checksums.env
	}
	if checksums.sharedEnv != "" {
		podAnnotations[pipeline.AnnotationChecksumSharedEnv] = checksums.sharedEnv
	}

	env := req.Capsule().Spec.Env
	envFrom := pipeline.EnvSources(env.From)
	if !env.DisableAutomatic {
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

	for i, emptyPath := range strings.Split(req.Capsule().GetAnnotations()[AnnotationEmptyDirs], ",") {
		if emptyPath == "" {
			continue
		}

		name := fmt.Sprintf("empty-dir-%d", i)
		volumes = append(volumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      name,
			MountPath: emptyPath,
		})
	}

	// As described in "Kubernetes in Action" by Marko Lukša, and elaborated here:
	// https://blog.gruntwork.io/delaying-shutdown-to-wait-for-pod-deletion-propagation-445f779a8304
	// there is a race in Kubernetes where a immediately terminated Pod (e.g. graceful shutdown
	// terminated within seconds), the kube-nodes may not be aware of the termination yet and may
	// still forward traffic to this pod. As a remedy, it's recommended to delay shutdown by 5-10
	// seconds.
	// This is accomplished by applying a 7-second sleep in a preStop hook as follow.
	// To avoid unneeded hooks, we only do it if the Capsule has any network interfaces associated
	// to it.
	// Note: Not all system may have the sleep command. In that case, a Kubernetes event may be raised
	// but the overall result is a no-op.
	var lc *v1.Lifecycle
	if len(req.Capsule().Spec.Interfaces) > 0 {
		lc = &v1.Lifecycle{
			PreStop: &v1.LifecycleHandler{
				Exec: &v1.ExecAction{
					Command: []string{"sleep", "10"},
				},
			},
		}
	}

	c := v1.Container{
		Name:    req.Capsule().Name,
		Image:   req.Capsule().Spec.Image,
		EnvFrom: envFrom,
		Env: []v1.EnvVar{
			{
				Name:  "RIG_CAPSULE_NAME",
				Value: req.Capsule().Name,
			},
		},
		VolumeMounts: volumeMounts,
		Resources:    makeResourceRequirements(req.Capsule()),
		Args:         req.Capsule().Spec.Args,
		Lifecycle:    lc,
	}

	if req.Capsule().Spec.Command != "" {
		c.Command = []string{req.Capsule().Spec.Command}
	}

	replicas := ptr.New(int32(req.Capsule().Spec.Scale.Horizontal.Instances.Min))
	hasHPA, err := p.shouldCreateHPA(req)
	if err != nil {
		return nil, err
	}
	if hasHPA && current != nil && current.Spec.Replicas != nil {
		cur := uint32(*current.Spec.Replicas)
		ins := req.Capsule().Spec.Scale.Horizontal.Instances
		if cur < ins.Min {
			cur = ins.Min
		} else if ins.Max != nil && cur > *ins.Max {
			cur = *ins.Max
		}
		replicas = ptr.New(int32(cur))
	}

	strategy := appsv1.RollingUpdateDeploymentStrategyType
	if ok, _ := strconv.ParseBool(req.Capsule().GetAnnotations()[AnnotationRecreateStrategy]); ok {
		strategy = appsv1.RecreateDeploymentStrategyType
	}

	d := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: p.getPodsSelector(current, req),
			},
			Replicas: replicas,
			Strategy: appsv1.DeploymentStrategy{
				Type: strategy,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: podAnnotations,
					Labels:      p.getPodLabels(current, req),
				},
				Spec: v1.PodSpec{
					Containers:   []v1.Container{c},
					Volumes:      volumes,
					NodeSelector: req.Capsule().Spec.NodeSelector,
				},
			},
		},
	}

	if cfgs.imagePullSecret != "" {
		d.Spec.Template.Spec.ImagePullSecrets = []v1.LocalObjectReference{{
			Name: cfgs.imagePullSecret,
		}}
	}

	return d, nil
}

func (p *Plugin) handleInterfaces(req pipeline.CapsuleRequest, deployment *appsv1.Deployment) error {
	for i, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name != req.Capsule().Name {
			continue
		}

		var ports []v1.ContainerPort
		for _, ni := range req.Capsule().Spec.Interfaces {
			ports = append(ports, v1.ContainerPort{
				Name:          ni.Name,
				ContainerPort: ni.Port,
			})

			if ni.Liveness != nil {
				container.LivenessProbe = createProbe(ni.Liveness, ni.Port)
			}

			if ni.Readiness != nil {
				container.ReadinessProbe = createProbe(ni.Readiness, ni.Port)
			}
		}
		container.Ports = ports
		deployment.Spec.Template.Spec.Containers[i] = container
	}

	return nil
}

func createProbe(probe *v1alpha2.InterfaceProbe, port int32) *v1.Probe {
	switch {
	case probe.Path != "":
		return &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: probe.Path,
					Port: intstr.FromInt32(port),
				},
			},
		}
	case probe.TCP:
		return &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				TCPSocket: &v1.TCPSocketAction{
					Port: intstr.FromInt32(port),
				},
			},
		}
	case probe.GRPC != nil && probe.GRPC.Enabled:
		return &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				GRPC: &v1.GRPCAction{
					Service: &probe.GRPC.Service,
					Port:    port,
				},
			},
		}
	default:
		return nil
	}
}

func (p *Plugin) createService(req pipeline.CapsuleRequest) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Capsule().Name,
			Namespace: req.Capsule().Namespace,
			Labels: map[string]string{
				pipeline.LabelCapsule: req.Capsule().Name,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				pipeline.LabelCapsule: req.Capsule().Name,
			},
		},
	}

	for _, inf := range req.Capsule().Spec.Interfaces {
		svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
			Name:       inf.Name,
			Port:       inf.Port,
			TargetPort: intstr.FromString(inf.Name),
		})
	}

	return svc
}

func (p *Plugin) getPodLabels(current *appsv1.Deployment, req pipeline.CapsuleRequest) map[string]string {
	labels := map[string]string{}
	maps.Copy(labels, p.getPodsSelector(current, req))
	labels[pipeline.LabelCapsule] = req.Capsule().Name
	return labels
}

func (p *Plugin) getPodsSelector(current *appsv1.Deployment, req pipeline.CapsuleRequest) map[string]string {
	if current != nil {
		if s := current.Spec.Selector; s != nil {
			if len(s.MatchLabels) > 0 && len(s.MatchExpressions) == 0 {
				return s.MatchLabels
			}
		}
	}

	return map[string]string{
		pipeline.LabelCapsule: req.Capsule().Name,
	}
}

func (p *Plugin) getConfigChecksums(req pipeline.CapsuleRequest, cfgs configs) (checksums, error) {
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

func configEnvChecksum(req pipeline.CapsuleRequest, cfgs configs) (string, error) {
	if len(req.Capsule().Spec.Env.From) == 0 {
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

func configFilesChecksum(req pipeline.CapsuleRequest, cfgs configs) (string, error) {
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

func (p *Plugin) getConfigs(ctx context.Context, req pipeline.CapsuleRequest) (*configs, error) {
	configs := &configs{
		configMaps: map[string]*v1.ConfigMap{},
		secrets:    map[string]*v1.Secret{},
	}

	// Get shared env
	var configMapList v1.ConfigMapList
	if err := req.Reader().List(ctx, &configMapList, &client.ListOptions{
		Namespace: req.Capsule().Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			pipeline.LabelSharedConfig: "true",
		}),
	}); err != nil {
		return nil, fmt.Errorf("could not list shared env configmaps: %w", err)
	}
	configs.sharedEnvConfigMaps = make([]string, len(configMapList.Items))
	for i, cm := range configMapList.Items {
		configs.sharedEnvConfigMaps[i] = cm.GetName()
		configs.configMaps[cm.Name] = &cm
	}
	var secretList v1.SecretList
	if err := req.Reader().List(ctx, &secretList, &client.ListOptions{
		Namespace: req.Capsule().Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			pipeline.LabelSharedConfig: "true",
		}),
	}); err != nil {
		return nil, fmt.Errorf("could not list shared env secrets: %w", err)
	}
	configs.sharedEnvSecrets = make([]string, len(secretList.Items))
	for i, s := range secretList.Items {
		configs.sharedEnvSecrets[i] = s.GetName()
		configs.secrets[s.Name] = &s
	}

	env := req.Capsule().Spec.Env

	// Get automatic env
	if !env.DisableAutomatic {
		if err := p.setUsedSource(ctx, req, configs, "ConfigMap", req.Capsule().Name, false); err != nil {
			return nil, err
		}

		if err := p.setUsedSource(ctx, req, configs, "Secret", req.Capsule().Name, false); err != nil {
			return nil, err
		}
	}

	// Get envs
	for _, e := range env.From {
		if err := p.setUsedSource(ctx, req, configs, e.Kind, e.Name, true); err != nil {
			return nil, err
		}
	}

	// Get files
	for _, f := range req.Capsule().Spec.Files {
		if err := p.setUsedSource(ctx, req, configs, f.Ref.Kind, f.Ref.Name, true); err != nil {
			return nil, err
		}
	}

	if secret := req.Capsule().Annotations[pipeline.AnnotationPullSecret]; secret != "" {
		configs.imagePullSecret = secret
		if err := p.setUsedSource(ctx, req, configs, "Secret", secret, true); err != nil {
			return nil, err
		}
	}

	return configs, nil
}

func (p *Plugin) setUsedSource(
	ctx context.Context,
	req pipeline.CapsuleRequest,
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

		_ = req.MarkUsedObject(ref)
	}()

	switch kind {
	case "ConfigMap":
		if _, ok := cfgs.configMaps[name]; ok {
			return nil
		}
		var cm v1.ConfigMap
		if err := req.Reader().Get(ctx, types.NamespacedName{
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
		if err := req.Reader().Get(ctx, types.NamespacedName{
			Name:      name,
			Namespace: req.Capsule().Namespace,
		}, &s); err != nil {
			return fmt.Errorf("could not get referenced environment secret: %w", err)
		}
		cfgs.secrets[s.Name] = &s
	}

	return nil
}

func (p *Plugin) shouldCreateHPA(req pipeline.CapsuleRequest) (bool, error) {
	_, res, err := p.createHPA(req)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (p *Plugin) createHPA(req pipeline.CapsuleRequest) (*autoscalingv2.HorizontalPodAutoscaler, bool, error) {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2",
		},
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
	imagePullSecret     string
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
