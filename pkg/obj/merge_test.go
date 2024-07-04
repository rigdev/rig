package obj

import (
	"testing"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func TestMerger(t *testing.T) {
	t.Parallel()

	testMerge(t, "test override",
		&v1alpha1.PlatformConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       "PlatformConfig",
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			Auth: v1alpha1.Auth{
				SSO: v1alpha1.SSO{
					OIDCProviders: map[string]v1alpha1.OIDCProvider{
						"test": {
							ClientSecret: "secret",
						},
					},
				},
			},
		},
		&v1alpha1.PlatformConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       "PlatformConfig",
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			Auth: v1alpha1.Auth{
				SSO: v1alpha1.SSO{
					OIDCProviders: map[string]v1alpha1.OIDCProvider{
						"test": {
							ClientID: "id",
						},
					},
				},
			},
		},
		&v1alpha1.PlatformConfig{
			TypeMeta: v1.TypeMeta{
				Kind:       "PlatformConfig",
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			Auth: v1alpha1.Auth{
				SSO: v1alpha1.SSO{
					OIDCProviders: map[string]v1alpha1.OIDCProvider{
						"test": {
							ClientID:     "id",
							ClientSecret: "secret",
						},
					},
				},
			},
		},
		&v1alpha1.PlatformConfig{},
	)

	podMeta := metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
	testMerge(t, "test container change",
		&corev1.Pod{
			TypeMeta: podMeta,
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "c2",
						Image:   "image2",
						Command: []string{"cmd2-new"},
						Args:    []string{"arg1, arg2"},
					},
				},
			},
		},
		&corev1.Pod{
			TypeMeta: podMeta,
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "c1",
						Image:   "image1",
						Command: []string{"cmd1"},
					},
					{
						Name:    "c2",
						Image:   "image2",
						Command: []string{"cmd2"},
					},
				},
			},
		},
		&corev1.Pod{
			TypeMeta: podMeta,
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "c1",
						Image:   "image1",
						Command: []string{"cmd1"},
					},
					{
						Name:    "c2",
						Image:   "image2",
						Command: []string{"cmd2-new"},
						Args:    []string{"arg1, arg2"},
					},
				},
			},
		},
		&corev1.Pod{},
	)
	testMerge(t, "test container add",
		&corev1.Pod{
			TypeMeta: podMeta,
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "c3",
						Image:   "image3",
						Command: []string{"cmd3"},
					},
				},
			},
		},
		&corev1.Pod{
			TypeMeta: podMeta,
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "c1",
						Image:   "image1",
						Command: []string{"cmd1"},
					},
					{
						Name:    "c2",
						Image:   "image2",
						Command: []string{"cmd2"},
					},
				},
			},
		},
		&corev1.Pod{
			TypeMeta: podMeta,
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "c3",
						Image:   "image3",
						Command: []string{"cmd3"},
					},
					{
						Name:    "c1",
						Image:   "image1",
						Command: []string{"cmd1"},
					},
					{
						Name:    "c2",
						Image:   "image2",
						Command: []string{"cmd2"},
					},
				},
			},
		},
		&corev1.Pod{},
	)
	testMerge(t, "test container merge",
		&corev1.Pod{
			TypeMeta: podMeta,
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "c1",
						Image: "image3",
						Ports: []corev1.ContainerPort{
							{
								Name:          "port1",
								ContainerPort: 1001,
								HostPort:      6969,
								HostIP:        "some-ip",
							},
							{
								Name:          "port3",
								HostPort:      5555,
								ContainerPort: 1003,
							},
							{
								Name:          "port4",
								HostPort:      4567,
								ContainerPort: 1004,
							},
						},
					},
				},
			},
		},
		&corev1.Pod{
			TypeMeta: podMeta,
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "c1",
						Image:   "image1",
						Command: []string{"cmd1"},
						Ports: []corev1.ContainerPort{
							{
								Name:          "port1",
								ContainerPort: 1001,
								HostPort:      8080,
								Protocol:      "protocol",
							},
							{
								Name:          "port2",
								ContainerPort: 1002,
								HostPort:      1234,
								HostIP:        "ip",
							},
						},
					},
					{
						Name:    "c2",
						Image:   "image2",
						Command: []string{"cmd2"},
					},
				},
			},
		},
		&corev1.Pod{
			TypeMeta: podMeta,
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "c1",
						Image:   "image3",
						Command: []string{"cmd1"},
						Ports: []corev1.ContainerPort{
							{
								Name:          "port1",
								HostPort:      6969,
								HostIP:        "some-ip",
								Protocol:      "protocol",
								ContainerPort: 1001,
							},
							{
								Name:          "port3",
								HostPort:      5555,
								ContainerPort: 1003,
							},
							{
								Name:          "port4",
								HostPort:      4567,
								ContainerPort: 1004,
							},
							{
								Name:          "port2",
								HostPort:      1234,
								HostIP:        "ip",
								ContainerPort: 1002,
							},
						},
					},
					{
						Name:    "c2",
						Image:   "image2",
						Command: []string{"cmd2"},
					},
				},
			},
		},
		&corev1.Pod{},
	)

	deploymentMeta := metav1.TypeMeta{
		Kind:       "Deployment",
		APIVersion: "apps/v1",
	}
	testMerge(t, "JSON omitEmpty behaviour",
		&appsv1.Deployment{
			TypeMeta: deploymentMeta,
			Spec: appsv1.DeploymentSpec{
				Replicas: ptr.New[int32](2),
			},
		},
		&appsv1.Deployment{
			TypeMeta: deploymentMeta,
			Spec: appsv1.DeploymentSpec{
				Replicas:        ptr.New[int32](1),
				MinReadySeconds: 1, // Not overwritten (has omitempty)
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{ // Overwritten (no omitEmpty)
							Image: "some-image",
						}},
						InitContainers: []corev1.Container{{ // Not overwritten (has omitEmpty)
							Image: "init-image",
						}},
					},
				},
				Selector: &v1.LabelSelector{ // Overwritten (no omitEmpty)
					MatchLabels: map[string]string{ // Has omitEmpty, but because parent field doesn't have, it gets overwritten
						"some-label": "some-value",
					},
				},
			},
		},
		&appsv1.Deployment{
			TypeMeta: deploymentMeta,
			Spec: appsv1.DeploymentSpec{
				Replicas:        ptr.New[int32](2),
				MinReadySeconds: 1,
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{{
							Image: "init-image",
						}},
					},
				},
			},
		},
		&appsv1.Deployment{},
	)
}

func testMerge[T runtime.Object](t *testing.T, name string, src, dst, expected, empty T) {
	codecs := serializer.NewCodecFactory(scheme.New())

	info, _ := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), runtime.ContentTypeJSON)
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		res, err := Merge(src, dst, empty, info.Serializer)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}
