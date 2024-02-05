package k8s_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//+kubebuilder:scaffold:imports
)

func (s *PluginTestSuite) TestObjectPluginDeploymentReplicas() {
	ctx := context.Background()
	nsName := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: "default",
	}

	s.by("Creating a capsule")

	capsule := v1alpha2.Capsule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Spec: v1alpha2.CapsuleSpec{
			Image: "nginx:1.25.1",
			Scale: v1alpha2.CapsuleScale{
				Horizontal: v1alpha2.HorizontalScale{
					Instances: v1alpha2.Instances{
						Min: uint32(1),
					},
				},
			},
		},
	}

	s.Require().NoError(s.Client.Create(ctx, &capsule))
	s.expectResources(ctx, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name: nsName.Name,
						}},
					},
				},
				// The 2 comes from the plugin.
				Replicas: ptr.New[int32](2),
			},
		},
	})
}

func (s *PluginTestSuite) TestSidecarPluginNginxContainer() {
	ctx := context.Background()
	nsName := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: "default",
	}

	s.by("Creating a capsule")

	capsule := v1alpha2.Capsule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Spec: v1alpha2.CapsuleSpec{
			Image: "nginx:1.25.1",
			Scale: v1alpha2.CapsuleScale{
				Horizontal: v1alpha2.HorizontalScale{
					Instances: v1alpha2.Instances{
						Min: uint32(1),
					},
				},
			},
		},
	}

	s.Require().NoError(s.Client.Create(ctx, &capsule))
	s.expectResources(ctx, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: nsName.Name,
							},
							// This sidecar container is added by the plugin.
							{
								Name:  "nginx",
								Image: "nginx",
							},
						},
					},
				},
			},
		},
	})
}

func (s *PluginTestSuite) TestInitContainerPluginStartupEcho() {
	ctx := context.Background()
	nsName := types.NamespacedName{
		Name:      uuid.NewString(),
		Namespace: "default",
	}

	s.by("Creating a capsule")

	capsule := v1alpha2.Capsule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Spec: v1alpha2.CapsuleSpec{
			Image: "nginx:1.25.1",
			Scale: v1alpha2.CapsuleScale{
				Horizontal: v1alpha2.HorizontalScale{
					Instances: v1alpha2.Instances{
						Min: uint32(1),
					},
				},
			},
		},
	}

	s.Require().NoError(s.Client.Create(ctx, &capsule))
	s.expectResources(ctx, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: nsName.Name,
							},
						},
						// This init container is added by the plugin.
						InitContainers: []v1.Container{
							{
								Image:   "alpine",
								Name:    "startup",
								Command: []string{"sh", "-c", "echo Hello"},
							},
						},
					},
				},
			},
		},
	})
}

func (s *PluginTestSuite) by(msg string) {
	s.T().Log("STEP: ", msg)
}
