package envmapping

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	Name                 = "rigdev.env_mapping"
	AnnotationEnvMapping = "plugin.rig.dev/env-mapping"
)

// Configuration for the env_mapping plugin
// +kubebuilder:object:root=true
type Config struct{}

type AnnotationValue struct {
	Sources []AnnotationSource `json:"sources"`
}

type AnnotationSource struct {
	// Container name default to capsule name.
	Container string `json:"container,omitempty"`
	// Optional ConfigMap reference.
	ConfigMap string `json:"configMap,omitempty"`
	// Optional Secret reference.
	Secret string `json:"secret,omitempty"`
	// Mappings ENV:KEY
	Mappings map[string]string `json:"mappings"`
}

type Plugin struct {
	plugin.NoWatchObjectStatus
}

func (p *Plugin) Initialize(_ plugin.InitializeRequest) error {
	return nil
}

func (p *Plugin) ComputeConfig(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) (string, error) {
	return plugin.ParseCapsuleTemplatedConfigToString[Config](nil, req)
}

func (p *Plugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	value, ok := req.Capsule().Annotations[AnnotationEnvMapping]
	if !ok {
		return nil
	}

	var data AnnotationValue
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	deployment := &appsv1.Deployment{}
	if err := req.GetNewInto(deployment); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, source := range data.Sources {
		container := source.Container
		if container == "" {
			container = req.Capsule().GetName()
		}

		for i, c := range deployment.Spec.Template.Spec.Containers {
			if c.Name != container {
				continue
			}

			for env, key := range source.Mappings {
				envVar := corev1.EnvVar{
					Name: env,
				}
				switch {
				case source.ConfigMap != "":
					envVar.ValueFrom = &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: source.ConfigMap,
							},
							Key: key,
						},
					}
					if err := req.MarkUsedObject(v1alpha2.UsedResource{
						Ref: &corev1.TypedLocalObjectReference{
							Kind: "ConfigMap",
							Name: source.ConfigMap,
						},
					}); err != nil {
						return err
					}
				case source.Secret != "":
					envVar.ValueFrom = &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: source.Secret,
							},
							Key: key,
						},
					}
					if err := req.MarkUsedObject(v1alpha2.UsedResource{
						Ref: &corev1.TypedLocalObjectReference{
							Kind: "Secret",
							Name: source.Secret,
						},
					}); err != nil {
						return err
					}
				}

				c.Env = append(c.Env, envVar)
			}

			deployment.Spec.Template.Spec.Containers[i] = c

			break
		}
	}

	return req.Set(deployment)
}
