package k8s_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	//+kubebuilder:scaffold:imports

	"github.com/rigdev/rig/pkg/ptr"
)

func (s *K8sTestSuite) TestController() {
	k8sClient := s.Client
	t := s.Suite.T()

	by(t, "Creating namespace")

	ctx := context.Background()
	nsName := types.NamespacedName{
		Name:      "test",
		Namespace: "nginx",
	}

	ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName.Namespace}}
	assert.NoError(t, k8sClient.Create(ctx, ns))

	by(t, "Creating a capsule")

	capsule := v1alpha1.Capsule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Spec: v1alpha1.CapsuleSpec{
			Image: "nginx:1.25.1",
			HorizontalScale: v1alpha1.HorizontalScale{
				MinReplicas: ptr.New(uint32(1)),
				MaxReplicas: ptr.New(uint32(1)),
			},
		},
	}

	assert.NoError(t, k8sClient.Create(ctx, &capsule))

	var deploy appsv1.Deployment
	assert.Eventually(t, func() bool {
		if err := k8sClient.Get(ctx, nsName, &deploy); err != nil {
			return false
		}
		return true
	}, waitFor, tick)

	if assert.Len(t, deploy.Spec.Template.Spec.Containers, 1) {
		assert.Equal(t, deploy.Spec.Template.Spec.Containers[0].Image, "nginx:1.25.1")
	}

	capsuleOwnerRef := metav1.OwnerReference{
		Kind:               "Capsule",
		APIVersion:         v1alpha1.GroupVersion.Identifier(),
		UID:                capsule.UID,
		Name:               nsName.Name,
		Controller:         ptr.New(true),
		BlockOwnerDeletion: ptr.New(true),
	}

	if assert.Len(t, deploy.OwnerReferences, 1) {
		assert.Equal(t, capsuleOwnerRef, deploy.OwnerReferences[0])
	}

	expectResources(ctx, t, k8sClient, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: ns.Name,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{},
					},
				},
			},
		},
	})

	err := k8sClient.Get(ctx, nsName, &v1.Service{})
	assert.True(t, kerrors.IsNotFound(err))

	assert.NoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(&capsule), &capsule))

	by(t, "Adding an interface")

	capsule.Spec.Interfaces = []v1alpha1.CapsuleInterface{
		{
			Name: "http",
			Port: 80,
		},
	}
	assert.NoError(t, k8sClient.Update(ctx, &capsule))

	expectResources(ctx, t, k8sClient, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: ns.Name,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name: "test",
							Ports: []v1.ContainerPort{{
								Name:          "http",
								ContainerPort: 80,
							}},
						}},
					},
				},
			},
		},
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: ns.Name,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromString("http"),
				}},
			},
		},
	})

	err = k8sClient.Get(ctx, nsName, &netv1.Ingress{})
	assert.True(t, kerrors.IsNotFound(err))

	assert.NoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(&capsule), &capsule))

	by(t, "Enabling ingress")

	capsule.Spec.Interfaces[0].Public = &v1alpha1.CapsulePublicInterface{
		Ingress: &v1alpha1.CapsuleInterfaceIngress{
			Host: "test.com",
		},
	}
	assert.NoError(t, k8sClient.Update(ctx, &capsule))

	pt := netv1.PathTypePrefix
	expectResources(ctx, t, k8sClient, []client.Object{
		&netv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: ns.Name,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: netv1.IngressSpec{
				Rules: []netv1.IngressRule{{
					Host: "test.com",
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{{
								Path:     "/",
								PathType: &pt,
								Backend: netv1.IngressBackend{
									Service: &netv1.IngressServiceBackend{
										Name: "test",
										Port: netv1.ServiceBackendPort{
											Name: "http",
										},
									},
								},
							}},
						},
					},
				}},
				TLS: []netv1.IngressTLS{{
					Hosts:      []string{"test.com"},
					SecretName: "test-tls",
				}},
			},
		},
		&cmv1.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: ns.Name,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: cmv1.CertificateSpec{
				SecretName: "test-tls",
				IssuerRef: cmmeta.ObjectReference{
					Kind: cmv1.ClusterIssuerKind,
					Name: "test",
				},
				DNSNames: []string{
					"test.com",
				},
			},
		},
	})

	assert.NoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(&capsule), &capsule))

	by(t, "Changing ingress to loadbalancer")

	capsule.Spec.Interfaces[0].Public = &v1alpha1.CapsulePublicInterface{
		LoadBalancer: &v1alpha1.CapsuleInterfaceLoadBalancer{
			Port: 1,
		},
	}
	assert.NoError(t, k8sClient.Update(ctx, &capsule))

	assert.Eventually(t, func() bool {
		if err := k8sClient.Get(ctx, nsName, &netv1.Ingress{}); err != nil {
			if kerrors.IsNotFound(err) {
				return true
			}
		}
		return false
	}, waitFor, tick)

	expectResources(ctx, t, k8sClient, []client.Object{
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-lb", nsName.Name),
				Namespace: ns.Name,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{{
					Name:       "http",
					Port:       1,
					TargetPort: intstr.FromString("http"),
				}},
				Type: v1.ServiceTypeLoadBalancer,
			},
		},
	})

	by(t, "Adding an environment variable configmap")

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Data: map[string]string{
			"TEST": "test",
		},
	}

	assert.NoError(t, k8sClient.Create(ctx, cm))
	hash := sha256.New()
	hash.Write([]byte("TEST"))
	hash.Write([]byte("test"))
	expectResources(ctx, t, k8sClient, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: ns.Name,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"rig.dev/config-checksum": fmt.Sprintf("%x", hash.Sum(nil)),
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name: "test",
							Ports: []v1.ContainerPort{{
								Name:          "http",
								ContainerPort: 80,
							}},
						}},
					},
				},
			},
		},
	})

	assert.NoError(t, k8sClient.Delete(ctx, cm))

	by(t, "Adding a configfile configmap")

	cm = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-files", nsName.Name),
			Namespace: nsName.Namespace,
		},
		Data: map[string]string{
			"test.yaml":           "test1",
			"not-referenced.yaml": "test",
		},
	}

	assert.NoError(t, k8sClient.Create(ctx, cm))
	assert.NoError(t, k8sClient.Get(ctx, client.ObjectKeyFromObject(&capsule), &capsule))

	capsule.Spec.Files = []v1alpha1.File{{
		Path: "/etc/test/test.yaml",
		ConfigMap: &v1alpha1.FileContentRef{
			Name: cm.GetName(),
			Key:  "test.yaml",
		},
	}}

	assert.NoError(t, k8sClient.Update(ctx, &capsule))

	hash = sha256.New()
	hash.Write([]byte("test.yaml"))
	hash.Write([]byte("test1"))
	expectResources(ctx, t, k8sClient, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: ns.Name,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"rig.dev/config-checksum": fmt.Sprintf("%x", hash.Sum(nil)),
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name: "test",
							Ports: []v1.ContainerPort{{
								Name:          "http",
								ContainerPort: 80,
							}},
						}},
					},
				},
			},
		},
	})

	cm.Data["test.yaml"] = "test2"

	assert.NoError(t, k8sClient.Update(ctx, cm))
	hash = sha256.New()
	hash.Write([]byte("test.yaml"))
	hash.Write([]byte("test2"))
	expectResources(ctx, t, k8sClient, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: ns.Name,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"rig.dev/config-checksum": fmt.Sprintf("%x", hash.Sum(nil)),
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{{
							Name: "test",
							Ports: []v1.ContainerPort{{
								Name:          "http",
								ContainerPort: 80,
							}},
						}},
					},
				},
			},
		},
	})

	by(t, "Deleting the capsule")

	assert.NoError(t, k8sClient.Delete(ctx, &capsule))
	assert.Eventually(t, func() bool {
		if err := k8sClient.Get(ctx, nsName, &capsule); err != nil {
			if kerrors.IsNotFound(err) {
				return true
			}
		}
		return false
	}, waitFor, tick)
}

func by(t *testing.T, msg string) {
	t.Log("STEP: ", msg)
}

func expectResources(ctx context.Context, t *testing.T, k8sClient client.Client, resources []client.Object) {
	for _, r := range resources {
		c := 0
		cp := r.DeepCopyObject().(client.Object)
		for {
			err := k8sClient.Get(ctx, client.ObjectKeyFromObject(r), cp)
			if kerrors.IsNotFound(err) {
				time.Sleep(100 * time.Millisecond)
				continue
			} else if err != nil {
				require.NoError(t, err)
			}

			// Clear this property.
			cp.SetCreationTimestamp(metav1.Time{})

			bs1, err := json.Marshal(r)
			require.NoError(t, err)

			bs2, err := json.Marshal(cp)
			require.NoError(t, err)

			opt := jsondiff.DefaultConsoleOptions()
			diff, change := jsondiff.Compare(bs2, bs1, &opt)

			c++
			if jsondiff.SupersetMatch == diff {
				break
			} else if c > 20 {
				require.Equal(t, jsondiff.SupersetMatch, diff, change)
			}

			time.Sleep(250 * time.Millisecond)
		}
	}
}
