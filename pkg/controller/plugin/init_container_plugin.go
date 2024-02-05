package plugin

import (
	"bufio"
	"bytes"
	"context"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type InitContainerPlugin struct {
	cfg *v1alpha1.InitContainerPlugin
}

func NewInitContainerPlugin(cfg *v1alpha1.InitContainerPlugin) Plugin {
	return &InitContainerPlugin{
		cfg: cfg,
	}
}

func (s *InitContainerPlugin) Run(_ context.Context, req pipeline.Request) error {
	r := yaml.NewYAMLToJSONDecoder(bufio.NewReader(bytes.NewReader([]byte(s.cfg.Container))))
	var c corev1.Container
	if err := r.Decode(&c); err != nil {
		return err
	}

	key := req.ObjectKey(pipeline.AppsDeploymentGVK)

	object := pipeline.Get[*appsv1.Deployment](req, key)
	if object == nil {
		return nil
	}

	object.Spec.Template.Spec.InitContainers = append(object.Spec.Template.Spec.InitContainers, c)

	req.Set(key, object)
	return nil
}
