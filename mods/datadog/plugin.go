// +groupName=plugins.rig.dev -- Only used for config doc generation
package datadog

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/mod"
	"github.com/rigdev/rig/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
)

const Name = "rigdev.datadog"

// Configuration for the datadog plugin
// +kubebuilder:object:root=true
type Config struct {
	// DontAddEnabledAnnotation toggles if the pods should have an annotation
	// allowing the Datadog Admission controller to modify them.
	DontAddEnabledAnnotation bool `json:"dontAddEnabledAnnotation,omitempty"`
	// LibraryTag defines configuration for which datadog libraries to inject into the pods.
	LibraryTag LibraryTag `json:"libraryTag,omitempty"`
	// UnifiedServiceTags configures the values for the Unified Service datadog tags.
	UnifiedServiceTags UnifiedServiceTags `json:"unifiedServiceTags,omitempty"`
}

// LibraryTag defines configuration for which datadog libraries to let the admission controller inject into the pods
// The admission controller will inject libraries from a container with the specified tag if the field is set.
type LibraryTag struct {
	// Tag of the Java library container
	Java string `json:"java,omitempty"`
	// Tag of the JavaScript library container
	JavaScript string `json:"javascript,omitempty"`
	// Tag of the Python library container
	Python string `json:"python,omitempty"`
	// Tag of the .NET library container
	NET string `json:"net,omitempty"`
	// Tag of the Ruby library container
	Ruby string `json:"ruby,omitempty"`
}

// UnifiedServiceTags configures the values of the Unified Service datadog tags on both Deployment and Pods
type UnifiedServiceTags struct {
	// The env tag
	Env string `json:"env,omitempty"`
	// The service tag
	Service string `json:"service,omitempty"`
	// The version tag
	Version string `json:"version,omitempty"`
}

type Plugin struct {
	configBytes []byte
}

func (d *Plugin) Initialize(req mod.InitializeRequest) error {
	d.configBytes = req.Config
	return nil
}

func (d *Plugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	config, err := mod.ParseTemplatedConfig[Config](d.configBytes, req, mod.CapsuleStep[Config])
	if err != nil {
		return err
	}

	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	if deployment.Labels == nil {
		deployment.Labels = map[string]string{}
	}
	if deployment.Spec.Template.Labels == nil {
		deployment.Spec.Template.Labels = map[string]string{}
	}
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = map[string]string{}
	}

	if !config.DontAddEnabledAnnotation {
		deployment.Spec.Template.Labels["admission.datadoghq.com/enabled"] = "true"
	}

	l := config.LibraryTag
	tags := map[string]string{
		"admission.datadoghq.com/java-lib.version":   l.Java,
		"admission.datadoghq.com/js-lib.version":     l.JavaScript,
		"admission.datadoghq.com/python-lib.version": l.Python,
		"admission.datadoghq.com/dotnet-lib.version": l.NET,
		"admission.datadoghq.com/ruby-lib.version":   l.Ruby,
	}
	annotations := deployment.Spec.Template.Annotations
	for k, v := range tags {
		if v == "" {
			continue
		}
		annotations[k] = v
	}

	u := config.UnifiedServiceTags
	tags = map[string]string{
		"tags.datadoghq.com/env":     u.Env,
		"tags.datadoghq.com/service": u.Service,
		"tags.datadoghq.com/version": u.Version,
	}
	labels1, labels2 := deployment.Labels, deployment.Spec.Template.Labels
	for k, v := range tags {
		if v == "" {
			continue
		}
		labels1[k] = v
		labels2[k] = v
	}

	return req.Set(deployment)
}
