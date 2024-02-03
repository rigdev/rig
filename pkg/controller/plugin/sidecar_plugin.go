package plugin

import (
	"bufio"
	"bytes"
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type SidecarPlugin struct {
	Container string `json:"container"`
}

func NewSidecarPlugin(config map[string]string) (Plugin, error) {
	p := &SidecarPlugin{}
	return p, mapstructure.Decode(config, p)
}

func (s *SidecarPlugin) Run(_ context.Context, req pipeline.Request) error {
	r := yaml.NewYAMLToJSONDecoder(bufio.NewReader(bytes.NewReader([]byte(s.Container))))
	var c corev1.Container
	if err := r.Decode(&c); err != nil {
		return err
	}

	key := req.ObjectKey(pipeline.AppsDeploymentGVK)

	object := pipeline.Get[*appsv1.Deployment](req, key)
	if object == nil {
		return nil
	}

	object.Spec.Template.Spec.Containers = append(object.Spec.Template.Spec.Containers, c)

	req.Set(key, object)
	return nil
}
