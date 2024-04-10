// +groupName=plugins.rig.dev -- Only used for config doc generation
package googlesqlproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/mod"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const Name = "rigdev.google_cloud_sql_auth_proxy"

// Configuration for the google_cloud_sql_auth_proxy plugin
// +kubebuilder:object:root=true
type Config struct {
	// The image running on the new container. Defaults to gcr.io/cloud-sql-connectors/cloud-sql-proxy
	Image string `json:"image,omitempty"`
	// The tag of the image
	Tag string `json:"tag,omitempty"`
	// Arguments to pass to the cloud sql proxy. These will be appended after the instance connection names.
	Args []string `json:"args,omitempty"`
	// A list of either ConfigMaps or Secrets which will be mounted in as environment variables to the container.
	// It's a reuse of the Capsule CRD
	EnvFromSource []v1alpha2.EnvReference `json:"envFromSource,omitempty"`
	// A list of environment variables to set in the container
	EnvVars []corev1.EnvVar `json:"envVars,omitempty"`
	// Files is a list of files to mount in the container. These can either be
	// based on ConfigMaps or Secrets.
	// It's a reuse of the Capsule CRD
	Files []v1alpha2.File `json:"files,omitempty"`
	// Resources defines how large the container request should be. Defaults to the Kubernetes defaults.
	Resources Resources `json:"resources,omitempty"`
	// The instance_connection_names passed to the cloud_sql_proxy.
	InstanceConnectionNames []string `json:"instanceConnectionNames,omitempty"`
}

// Resources configures the size of the request of the cloud_sql_proxy container
type Resources struct {
	// The number of CPU cores to request.
	CPU string `json:"cpu"`
	// The bytes of memory to request.
	Memory string `json:"memory"`
}

type Plugin struct {
	configBytes []byte
}

func (p *Plugin) Initialize(req mod.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	config, err := mod.ParseTemplatedConfig[Config](p.configBytes, req, mod.CapsuleStep[Config])
	if err != nil {
		return err
	}
	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	var allnames []string
	for _, c := range deployment.Spec.Template.Spec.Containers {
		allnames = append(allnames, c.Name)
	}
	for _, c := range deployment.Spec.Template.Spec.InitContainers {
		allnames = append(allnames, c.Name)
	}
	for _, name := range allnames {
		if name == "google-cloud-sql-proxy" {
			return errors.New("there was already a container named 'google-cloud-sql-proxy'")
		}
	}

	image := config.Image
	if image == "" {
		image = "gcr.io/cloud-sql-connectors/cloud-sql-proxy"
	}
	if config.Tag != "" {
		image = fmt.Sprintf("%s:%s", image, config.Tag)
	}

	var args []string
	if len(config.InstanceConnectionNames) == 0 {
		return errors.New("instanceConnectionName was not given")
	}
	args = append(args, config.InstanceConnectionNames...)
	args = append(args, config.Args...)

	resources := map[corev1.ResourceName]resource.Quantity{}
	if config.Resources.CPU != "" {
		cpu, err := resource.ParseQuantity(config.Resources.CPU)
		if err != nil {
			return fmt.Errorf("cpu was malformed: %q", err)
		}
		resources[corev1.ResourceCPU] = cpu
	}
	if config.Resources.Memory != "" {
		memory, err := resource.ParseQuantity(config.Resources.Memory)
		if err != nil {
			return fmt.Errorf("memory was malformed: %q", err)
		}
		resources[corev1.ResourceMemory] = memory
	}

	volume, mounts := controller.FilesToVolumes(config.Files)
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
		EnvFrom: controller.EnvSources(config.EnvFromSource),
		Env:     config.EnvVars,
		Resources: corev1.ResourceRequirements{
			Requests: resources,
		},
		VolumeMounts: mounts,
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: ptr.New(true),
		},
		RestartPolicy: ptr.New(corev1.ContainerRestartPolicyAlways),
	}
	deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers, container)

	return req.Set(deployment)
}
