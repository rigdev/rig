// +groupName=plugins.rig.dev -- Only used for config doc generation
//
//nolint:revive
package service_account

import (
	"context"
	"maps"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	Name = "rigdev.service_account"
)

// Configuration for the deployment plugin
// +kubebuilder:object:root=true
type Config struct {
	// Name of the service-account to generated. Supports templating, e.g.
	//	`{{ .capsule.metadata.name }}-svcacc`
	Name string `json:"name"`

	// UseExisting will, if enabled, skip the creation of the service-account but
	// instead use an existing one, if it exists.
	UseExisting bool `json:"useExisting"`

	// Annotations to be added to all service accounts created.
	Annotations map[string]string `json:"annotations,omitempty"`
}

type Plugin struct {
	plugin.NoWatchObjectStatus

	configBytes []byte
}

func (p *Plugin) ComputeConfig(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) (string, error) {
	return plugin.ParseCapsuleTemplatedConfigToString[Config](p.configBytes, req)
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	var config Config
	var err error
	if len(p.configBytes) > 0 {
		config, err = plugin.ParseTemplatedConfig(p.configBytes, req, plugin.CapsuleStep[Config])
		if err != nil {
			return err
		}
	}

	name := config.Name
	if name == "" {
		name = req.Capsule().Name
	}

	var sa *corev1.ServiceAccount
	if config.UseExisting {
		sa = &corev1.ServiceAccount{}
		if err := errors.FromK8sClient(req.Reader().Get(ctx, types.NamespacedName{
			Namespace: req.Capsule().Namespace,
			Name:      name,
		}, sa)); errors.IsNotFound(err) {
			sa = nil
		} else if err != nil {
			return err
		}
	} else {
		sa = p.createServiceAccount(req, name, config.Annotations)

		if err := req.Set(sa); err != nil {
			return err
		}
	}

	if sa == nil {
		return nil
	}

	deploy := &appsv1.Deployment{}
	if err := req.GetNewInto(deploy); err != nil {
		return err
	}

	deploy.Spec.Template.Spec.ServiceAccountName = name

	return req.Set(deploy)
}

func (s *Plugin) createServiceAccount(
	req pipeline.CapsuleRequest, name string, annotations map[string]string,
) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   req.Capsule().Namespace,
			Annotations: maps.Clone(annotations),
		},
	}

	return sa
}
