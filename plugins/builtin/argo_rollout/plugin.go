// +groupName=plugins.rig.dev -- Only used for config doc generation
//
//nolint:revive
package argorollout

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "rigdev.argo_rollout"

	AnnotationRecreateStrategy = "rig.dev/recreate-strategy"
	AnnotationEmptyDirs        = "rig.dev/empty-dirs"
)

// Configuration for the argo_rollout plugin
// +kubebuilder:object:root=true
type Config struct {
	Strategy *v1alpha1.RolloutStrategy `json:"strategy,omitempty"`
}

type Plugin struct {
	configBytes []byte
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	if err := v1alpha1.AddToScheme(req.Scheme()); err != nil {
		return err
	}
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) ComputeConfig(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) (string, error) {
	return plugin.ParseCapsuleTemplatedConfigToString[Config](p.configBytes, req)
}

func (p *Plugin) Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	config, err := plugin.ParseCapsuleTemplatedConfig[Config](p.configBytes, req)
	if err != nil {
		return err
	}

	if config.Strategy == nil {
		return nil
	}

	deployment := &appsv1.Deployment{}
	if err := req.GetNewInto(deployment); err != nil {
		return fmt.Errorf("failed to GetNewInto: %w", err)
	}

	rollout := &v1alpha1.Rollout{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Rollout",
			APIVersion: "argoproj.io/v1alpha1",
		},
		Spec: v1alpha1.RolloutSpec{
			// We cannot set 'ResolvedFromRef' to true and we have to add an empty container list.
			// Setting the ResolvedFromRefs to true changes the marshaller of Rollout which empties the
			// Template.
			// Due to the json tags, this results in
			// {"template": {"spec": {"containers": null}}}
			// marshalled json. If the `"containers": null` is present, the validator fails because the field
			// needs a value set.
			// Doing this changes the json to
			// {"template": {"spec": {"containers": []}}}
			// and the validator no longer fails.
			TemplateResolvedFromRef: false,
			SelectorResolvedFromRef: false,
			Replicas:                deployment.Spec.Replicas,
			WorkloadRef: &v1alpha1.ObjectRef{
				APIVersion: deployment.APIVersion,
				Kind:       deployment.Kind,
				Name:       deployment.Name,
				ScaleDown:  "progressively",
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
			Strategy: *config.Strategy,
		},
	}
	// The rollout scales the deployment down to 0. If we havent set it to 0 ourselves, the Rollout and
	// Capsule controller goes into an infinite loop of reconcillig.
	deployment.Spec.Replicas = ptr.New[int32](0)
	if err := req.Set(deployment); err != nil {
		return err
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2",
		},
	}
	if err := req.GetNewInto(hpa); errors.IsNotFound(err) {
	} else if err != nil {
		return err
	} else {
		hpa.Spec.ScaleTargetRef = autoscalingv2.CrossVersionObjectReference{
			Kind:       rollout.Kind,
			APIVersion: rollout.APIVersion,
			Name:       req.Capsule().Name,
		}
		if err := req.Set(hpa); err != nil {
			return err
		}
	}

	if err := req.Set(rollout); err != nil {
		return err
	}

	return nil
}
