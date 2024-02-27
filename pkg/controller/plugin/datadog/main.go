package main

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
)

type Config struct {
	DontAddEnabledAnnotation bool               `json:"dontAddEnabledAnnotation,omitempty"`
	LibraryTag               LibraryTag         `json:"libraryTag,omitempty"`
	UnifiedServiceTags       UnifiedServiceTags `json:"unifiedServiceTags,omitempty"`
}

type LibraryTag struct {
	Java       string `json:"java,omitempty"`
	JavaScript string `json:"javascript,omitempty"`
	Python     string `json:"python,omitempty"`
	NET        string `json:"net,omitempty"`
	Ruby       string `json:"ruby,omitempty"`
}

type UnifiedServiceTags struct {
	Env     string `json:"env,omitempty"`
	Service string `json:"service,omitempty"`
	Version string `json:"version,omitempty"`
}

type datadogPlugin struct {
	config Config
}

func (p *datadogPlugin) LoadConfig(data []byte) error {
	return plugin.LoadYAMLConfig(data, &p.config)
}

func (p *datadogPlugin) Run(_ context.Context, req pipeline.CapsuleRequest, _ hclog.Logger) error {
	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	templateContext := plugin.NewTemplateContext()
	if err := templateContext.AddData("capsule", req.Capsule()); err != nil {
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

	if !p.config.DontAddEnabledAnnotation {
		deployment.Spec.Template.Labels["admission.datadoghq.com/enabled"] = "true"
	}

	l := p.config.LibraryTag
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
		v, err := templateContext.Parse(v)
		if err != nil {
			return err
		}
		annotations[k] = v
	}

	u := p.config.UnifiedServiceTags
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
		v, err := templateContext.Parse(v)
		if err != nil {
			return err
		}
		labels1[k] = v
		labels2[k] = v
	}

	return req.Set(deployment)
}

func main() {
	plugin.StartPlugin("datadog", &datadogPlugin{})
}
