package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Config struct {
	Image                   string                  `json:"image,omitempty"`
	Tag                     string                  `json:"tag,omitempty"`
	Args                    []string                `json:"args,omitempty"`
	EnvFromSource           []v1alpha2.EnvReference `json:"envFromSource,omitempty"`
	EnvVars                 []corev1.EnvVar         `json:"envVars,omitempty"`
	Files                   []v1alpha2.File         `json:"files,omitempty"`
	Resources               Resources               `json:"resources,omitempty"`
	InstanceConnectionNames []string                `json:"instanceConnectionNames,omitempty"`
}

type Resources struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type pluginParent struct {
	configBytes []byte
}

func (p *pluginParent) LoadConfig(data []byte) error {
	p.configBytes = data
	return nil
}

func (p *pluginParent) Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	config, err := plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep[Config])
	if err != nil {
		return err
	}
	pp := &pluginImpl{
		config: config,
	}
	return pp.run(ctx, req, logger)
}

type pluginImpl struct {
	config Config
}

func (p *pluginImpl) run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == "google-cloud-sql-proxy" {
			return errors.New("there was already a container named 'google-cloud-sql-proxy'")
		}
	}

	image := p.config.Image
	if image == "" {
		image = "gcr.io/cloud-sql-connectors/cloud-sql-proxy"
	}
	if p.config.Tag != "" {
		image = fmt.Sprintf("%s:%s", image, p.config.Tag)
	}

	var args []string
	if len(p.config.InstanceConnectionNames) == 0 {
		return errors.New("instanceConnectionName was not given")
	}
	args = append(args, p.config.InstanceConnectionNames...)
	args = append(args, p.config.Args...)

	resources := map[corev1.ResourceName]resource.Quantity{}
	if p.config.Resources.CPU != "" {
		cpu, err := resource.ParseQuantity(p.config.Resources.CPU)
		if err != nil {
			return fmt.Errorf("cpu was malformed: %q", err)
		}
		resources[corev1.ResourceCPU] = cpu
	}
	if p.config.Resources.Memory != "" {
		memory, err := resource.ParseQuantity(p.config.Resources.Memory)
		if err != nil {
			return fmt.Errorf("memory was malformed: %q", err)
		}
		resources[corev1.ResourceMemory] = memory
	}

	volume, mounts := controller.FilesToVolumes(p.config.Files)
	for _, v := range volume {
		for _, vv := range deployment.Spec.Template.Spec.Volumes {
			found := false
			if v.Name == vv.Name {
				found = true
				break
			}
			if !found {
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, v)
			}
		}
	}

	container := corev1.Container{
		Name:    "google-cloud-sql-proxy",
		Image:   image,
		Args:    args,
		EnvFrom: controller.EnvSources(p.config.EnvFromSource),
		Env:     p.config.EnvVars,
		Resources: corev1.ResourceRequirements{
			Requests: resources,
		},
		VolumeMounts: mounts,
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: ptr.New(true),
		},
	}
	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, container)

	return req.Set(deployment)
}

func main() {
	plugin.StartPlugin("rigdev.google_sql_proxy", &pluginParent{})
}
