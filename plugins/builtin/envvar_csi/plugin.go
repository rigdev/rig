package envvarcsi

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/ptr"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	csiv1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)

const (
	Name = "rigdev.envvar_csi"
)

// Configuration for the env_mapping plugin
// +kubebuilder:object:root=true
type Config struct {
	Provider      string `json:"provider,omitempty"`
	ContainerName string `json:"containerName,omitempty"`
}

type Plugin struct {
	plugin.NoWatchObjectStatus

	configBytes []byte
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) ComputeConfig(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) (string, error) {
	return plugin.ParseCapsuleTemplatedConfigToString[Config](p.configBytes, req)
}

func (p *Plugin) Run(ctx context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	config, err := plugin.ParseCapsuleTemplatedConfig[Config](p.configBytes, req)
	if err != nil {
		return err
	}
	if config.ContainerName == "" {
		config.ContainerName = req.Capsule().Name
	}

	deployment := &appsv1.Deployment{}
	if err := req.GetNewInto(deployment); err != nil {
		return err
	}
	var container *corev1.Container
	for idx, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == config.ContainerName {
			container = &deployment.Spec.Template.Spec.Containers[idx]
			break
		}
	}
	if container == nil {
		return nil
	}

	envvars, err := readEnvironmentVariables(ctx, container, req)
	if err != nil {
		return fmt.Errorf("failed to read env vars: %w", err)
	}

	var provider *csiv1.SecretProviderClass
	switch config.Provider {
	case "aws":
		provider, err = handleAWS(req, envvars)
	default:
		return fmt.Errorf("unexpected provider type '%s'", config.Provider)
	}
	if err != nil {
		return err
	}

	if provider == nil {
		return nil
	}

	data, err := obj.EncodeAny(provider)
	if err != nil {
		return err
	}
	object := &unstructured.Unstructured{}
	if err := obj.DecodeInto(data, object, req.Scheme()); err != nil {
		return err
	}
	if err := req.Set(object); err != nil {
		return err
	}

	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: fmt.Sprintf("csi"),
		VolumeSource: corev1.VolumeSource{
			CSI: &corev1.CSIVolumeSource{
				Driver:   "secrets-store.csi.k8s.io",
				ReadOnly: ptr.New(true),
				VolumeAttributes: map[string]string{
					"secretProviderClass": req.Capsule().Name,
				},
			},
		},
	})
	container.EnvFrom = append(container.EnvFrom, corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: k8sSecretName(req.Capsule().Name),
			},
		},
	})

	if err := req.Set(deployment); err != nil {
		return err
	}

	return nil
}

func handleAWS(req pipeline.CapsuleRequest, envvars map[string]string) (*csiv1.SecretProviderClass, error) {
	capName := req.Capsule().Name
	provider := &csiv1.SecretProviderClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SecretProviderClass",
			APIVersion: "secrets-store.csi.x-k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      capName,
			Namespace: req.Capsule().Namespace,
		},
		Spec: csiv1.SecretProviderClassSpec{
			Provider:   "aws",
			Parameters: map[string]string{},
		},
	}
	secretObject := &csiv1.SecretObject{
		SecretName: k8sSecretName(capName),
		Type:       "Opaque",
	}

	var objects []map[string]string

	for k, v := range envvars {
		var objectType string
		if strings.HasPrefix(v, "__ssmParameter__=") {
			v = strings.TrimPrefix(v, "__ssmParameter__=")
			objectType = "ssmparameter"
		} else if strings.HasPrefix(v, "__secretName__=") {
			v = strings.TrimPrefix(v, "__secretName__=")
			objectType = "secretsmanager"
		}
		if objectType != "" {
			obj := map[string]string{
				"objectName": v,
				"objectType": objectType,
			}
			objects = append(objects, obj)
			secretObject.Data = append(secretObject.Data, &csiv1.SecretObjectData{
				ObjectName: strings.ReplaceAll(v, "/", "_"),
				Key:        k,
			})
		}
	}

	if objects == nil {
		return nil, nil
	}

	objectsBytes, err := yaml.Marshal(objects)
	if err != nil {
		return nil, err
	}
	provider.Spec.Parameters["objects"] = string(objectsBytes)
	provider.Spec.SecretObjects = append(provider.Spec.SecretObjects, secretObject)

	return provider, nil
}

func k8sSecretName(capName string) string {
	return fmt.Sprintf("csi-envvars-%s", capName)
}

// TODO: Only reads from the Platform created ConfigMap containing the environment variables from the Raw field in the platform Capsule.
func readEnvironmentVariables(ctx context.Context, container *corev1.Container, req pipeline.CapsuleRequest) (map[string]string, error) {
	c := req.Capsule()
	configMap := &corev1.ConfigMap{}
	if err := req.Reader().Get(ctx, client.ObjectKey{
		Namespace: c.Namespace,
		Name:      c.Name,
	}, configMap); err != nil {
		return nil, err
	}
	return maps.Clone(configMap.Data), nil
}
