package controller

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/roclient"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReusePodSelectors(t *testing.T) {
	current := &appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "foobar",
			Namespace: "my-ns",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": "foobar",
				},
			},
		},
	}

	r := roclient.NewReader(scheme.New())
	require.NoError(t, r.AddObject(current))

	p := pipeline.New(nil, r, &v1alpha1.OperatorConfig{}, scheme.New(), logr.Discard())
	req := pipeline.NewCapsuleRequest(p, &v1alpha2.Capsule{ObjectMeta: v1.ObjectMeta{
		Name:      "foobar",
		Namespace: "my-ns",
	}, Status: &v1alpha2.CapsuleStatus{
		OwnedResources: []v1alpha2.OwnedResource{
			{
				Ref: &corev1.TypedLocalObjectReference{
					APIGroup: ptr.New(current.GetObjectKind().GroupVersionKind().Group),
					Kind:     current.Kind,
					Name:     current.Name,
				},
			},
		},
	}})

	s := NewDeploymentStep()
	p.AddStep(s)
	_, err := p.RunSteps(context.Background(), req, false)
	require.NoError(t, err)

	dep := &appsv1.Deployment{}
	require.NoError(t, req.GetNew(dep))
	require.Equal(t, map[string]string{
		"app.kubernetes.io/name": "foobar",
	}, dep.Spec.Selector.MatchLabels)
}
