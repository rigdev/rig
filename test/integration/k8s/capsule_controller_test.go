package k8s_test

import (
	"context"
	"crypto/sha256"
	"fmt"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/google/uuid"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/hash"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//+kubebuilder:scaffold:imports

	"github.com/rigdev/rig/pkg/ptr"
)

var nsName types.NamespacedName

func (s *K8sTestSuite) TestControllerSharedSecrets() {
	ctx := context.Background()
	nsName = types.NamespacedName{
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
						Max: ptr.New(uint32(1)),
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
			},
		},
	})

	s.by("Creating a namespace capsule environment secret")

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.NewString(),
			Namespace: nsName.Namespace,
			Labels: map[string]string{
				controller.LabelSharedConfig: "true",
			},
		},
		Data: map[string][]byte{
			"SECRET": []byte("secret"),
		},
	}

	s.Require().NoError(s.Client.Create(ctx, &secret))

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
							EnvFrom: []v1.EnvFromSource{
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: secret.Name,
										},
									},
								},
							},
						}},
					},
				},
			},
		},
	})

	s.by("Creating a specific capsule environment secret")

	s.Require().NoError(s.Client.Delete(ctx, &secret))
	s.Require().Eventually(func() bool {
		return kerrors.IsNotFound(
			s.Client.Get(ctx, client.ObjectKeyFromObject(&secret), &v1.Secret{}),
		)
	}, waitFor, tick)

	secret.Name = uuid.NewString()
	delete(secret.Labels, controller.LabelSharedConfig)
	secret.ResourceVersion = ""
	s.Require().NoError(s.Client.Create(ctx, &secret))

	s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.Env = v1alpha2.Env{
			From: []v1alpha2.EnvReference{
				{
					Kind: "Secret",
					Name: secret.Name,
				},
			},
		}
	})

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
							EnvFrom: []v1.EnvFromSource{
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: secret.Name,
										},
									},
								},
							},
						}},
					},
				},
			},
		},
	})

	s.by("Creating an auto env secret")

	autoSecret := secret.DeepCopy()
	autoSecret.Name = nsName.Name
	autoSecret.ResourceVersion = ""
	s.Require().NoError(s.Client.Create(ctx, autoSecret))

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
							EnvFrom: []v1.EnvFromSource{
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: secret.Name,
										},
									},
								},
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: nsName.Name,
										},
									},
								},
							},
						}},
					},
				},
			},
		},
	})

	s.by("Disabling automatic env")

	s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.Env.DisableAutomatic = true
	})

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
							EnvFrom: []v1.EnvFromSource{
								{
									SecretRef: &v1.SecretEnvSource{
										LocalObjectReference: v1.LocalObjectReference{
											Name: secret.Name,
										},
									},
								},
							},
						}},
					},
				},
			},
		},
	})
}

func (s *K8sTestSuite) TestController() {
	nsName = types.NamespacedName{
		Name:      "test",
		Namespace: "nginx",
	}
	ctx := context.Background()
	s.testCreateCapsule(ctx)
	s.testInterface(ctx)
	s.testIngress(ctx)
	s.testLoadbalancer(ctx)
	s.testEnvVar(ctx)
	s.testConfigMap(ctx)
	s.testHPA(ctx)
	s.testCronJob(ctx)
	s.testPrometheusServiceMonitor(ctx)
	s.testDeleteCapsule(ctx)
}

func (s *K8sTestSuite) getCapsule(ctx context.Context) (v1alpha2.Capsule, metav1.OwnerReference) {
	var capsule v1alpha2.Capsule
	s.Assert().NoError(s.Client.Get(ctx, nsName, &capsule))
	return capsule, metav1.OwnerReference{
		Kind:               "Capsule",
		APIVersion:         v1alpha2.GroupVersion.Identifier(),
		UID:                capsule.UID,
		Name:               nsName.Name,
		Controller:         ptr.New(true),
		BlockOwnerDeletion: ptr.New(true),
	}
}

func (s *K8sTestSuite) updateCapsule(ctx context.Context, update func(*v1alpha2.Capsule)) metav1.OwnerReference {
	capsule := &v1alpha2.Capsule{}
	s.Require().NoError(s.Client.Get(ctx, nsName, capsule))
	for {
		update(capsule)
		err := s.Client.Update(ctx, capsule)
		if kerrors.IsConflict(err) {
			s.Require().NoError(s.Client.Get(ctx, nsName, capsule))
			continue
		}
		s.Assert().NoError(err)
		return metav1.OwnerReference{
			Kind:               "Capsule",
			APIVersion:         v1alpha2.GroupVersion.Identifier(),
			UID:                capsule.UID,
			Name:               nsName.Name,
			Controller:         ptr.New(true),
			BlockOwnerDeletion: ptr.New(true),
		}
	}
}

func (s *K8sTestSuite) testCreateCapsule(ctx context.Context) {
	s.by("Creating namespace")
	ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName.Namespace}}
	s.Assert().NoError(s.Client.Create(ctx, ns))

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
						Max: ptr.New(uint32(1)),
					},
				},
			},
		},
	}

	s.Assert().NoError(s.Client.Create(ctx, &capsule))

	var deploy appsv1.Deployment
	s.Assert().Eventually(func() bool {
		if err := s.Client.Get(ctx, nsName, &deploy); err != nil {
			return false
		}
		return true
	}, waitFor, tick)

	if s.Assert().Len(deploy.Spec.Template.Spec.Containers, 1) {
		s.Assert().Equal(deploy.Spec.Template.Spec.Containers[0].Image, "nginx:1.25.1")
	}

	capsuleOwnerRef := metav1.OwnerReference{
		Kind:               "Capsule",
		APIVersion:         v1alpha2.GroupVersion.Identifier(),
		UID:                capsule.UID,
		Name:               nsName.Name,
		Controller:         ptr.New(true),
		BlockOwnerDeletion: ptr.New(true),
	}

	if s.Assert().Len(deploy.OwnerReferences, 1) {
		s.Assert().Equal(capsuleOwnerRef, deploy.OwnerReferences[0])
	}

	s.expectResources(ctx, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
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

	err := s.Client.Get(ctx, nsName, &v1.Service{})
	s.Assert().True(kerrors.IsNotFound(err))
}

func (s *K8sTestSuite) testInterface(ctx context.Context) {
	s.by("Adding an interface")
	capsuleOwnerRef := s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.Interfaces = []v1alpha2.CapsuleInterface{
			{
				Name: "http",
				Port: 80,
			},
		}
	})

	s.expectResources(ctx, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
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
				Namespace: nsName.Namespace,
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

	err := s.Client.Get(ctx, nsName, &netv1.Ingress{})
	s.Assert().True(kerrors.IsNotFound(err))
}

func (s *K8sTestSuite) testIngress(ctx context.Context) {
	s.by("Enabling ingress")
	capsuleOwnerRef := s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.Interfaces[0].Public = &v1alpha2.CapsulePublicInterface{
			Ingress: &v1alpha2.CapsuleInterfaceIngress{
				Host: "test.com",
			},
		}
	})

	pt := netv1.PathTypeExact
	s.expectResources(ctx, []client.Object{
		&netv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
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
				Namespace: nsName.Namespace,
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

	capsuleOwnerRef = s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.Interfaces[0].Public.Ingress.Paths = []string{"/test1", "/test2"}
	})

	s.expectResources(ctx, []client.Object{
		&netv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: netv1.IngressSpec{
				Rules: []netv1.IngressRule{{
					Host: "test.com",
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path:     "/test1",
									PathType: &pt,
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: "test",
											Port: netv1.ServiceBackendPort{
												Name: "http",
											},
										},
									},
								},
								{
									Path:     "/test2",
									PathType: &pt,
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: "test",
											Port: netv1.ServiceBackendPort{
												Name: "http",
											},
										},
									},
								},
							},
						},
					},
				}},
				TLS: []netv1.IngressTLS{{
					Hosts:      []string{"test.com"},
					SecretName: "test-tls",
				}},
			},
		},
	})
}

func (s *K8sTestSuite) testLoadbalancer(ctx context.Context) {
	s.by("Changing ingress to loadbalancer")
	capsuleOwnerRef := s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.Interfaces[0].Public = &v1alpha2.CapsulePublicInterface{
			LoadBalancer: &v1alpha2.CapsuleInterfaceLoadBalancer{
				Port: 1,
			},
		}
	})

	s.Assert().Eventually(func() bool {
		if err := s.Client.Get(ctx, nsName, &netv1.Ingress{}); err != nil {
			if kerrors.IsNotFound(err) {
				return true
			}
		}
		return false
	}, waitFor, tick)

	s.expectResources(ctx, []client.Object{
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-lb", nsName.Name),
				Namespace: nsName.Namespace,
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
}

func (s *K8sTestSuite) testEnvVar(ctx context.Context) {
	s.by("Adding an environment variable configmap")
	_, capsuleOwnerRef := s.getCapsule(ctx)
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Data: map[string]string{
			"TEST": "test",
		},
	}

	s.Require().NoError(s.Client.Create(ctx, cm))

	h := sha256.New()
	s.Require().NoError(hash.ConfigMap(h, cm))

	s.expectResources(ctx, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							controller.AnnotationChecksumAutoEnv: fmt.Sprintf("%x", h.Sum(nil)),
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

	s.Assert().NoError(s.Client.Delete(ctx, cm))
}

func (s *K8sTestSuite) testConfigMap(ctx context.Context) {
	s.by("Adding a configfile configmap")

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-files", nsName.Name),
			Namespace: nsName.Namespace,
		},
		Data: map[string]string{
			"test.yaml":           "test1",
			"not-referenced.yaml": "test",
		},
	}

	s.Assert().NoError(s.Client.Create(ctx, cm))

	capsuleOwnerRef := s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.Files = []v1alpha2.File{{
			Path: "/etc/test/test.yaml",
			Ref: &v1alpha2.FileContentReference{
				Kind: "ConfigMap",
				Name: cm.GetName(),
				Key:  "test.yaml",
			},
		}}
	})

	h := sha256.New()
	s.Assert().NoError(hash.ConfigMapKeys(h, []string{"test.yaml"}, cm))
	s.expectResources(ctx, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							controller.AnnotationChecksumFiles: fmt.Sprintf("%x", h.Sum(nil)),
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

	s.Assert().NoError(s.Client.Update(ctx, cm))
	h = sha256.New()
	s.Assert().NoError(hash.ConfigMapKeys(h, []string{"test.yaml"}, cm))
	s.expectResources(ctx, []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							controller.AnnotationChecksumFiles: fmt.Sprintf("%x", h.Sum(nil)),
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

	err := s.Client.Get(ctx, nsName, &autoscalingv2.HorizontalPodAutoscaler{})
	s.Assert().True(kerrors.IsNotFound(err))
	s.getCapsule(ctx)
}

func (s *K8sTestSuite) testHPA(ctx context.Context) {
	s.by("Adding an HPA")
	capsuleOwnerRef := s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.Scale.Horizontal.CPUTarget = &v1alpha2.CPUTarget{
			Utilization: ptr.New(uint32(80)),
		}
		c.Spec.Scale.Horizontal.CustomMetrics = []v1alpha2.CustomMetric{
			{
				InstanceMetric: &v1alpha2.InstanceMetric{
					MetricName:   "some-metric",
					AverageValue: "10",
				},
			},
			{
				ObjectMetric: &v1alpha2.ObjectMetric{
					MetricName: "object-metric",
					Value:      "5",
					DescribedObject: autoscalingv2.CrossVersionObjectReference{
						Kind: "Service",
						Name: "my-service",
					},
				},
			},
		}
		c.Spec.Scale.Horizontal.Instances.Max = ptr.New(uint32(3))
	})

	s.expectResources(ctx, []client.Object{
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					capsuleOwnerRef,
				},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind:       "Deployment",
					Name:       nsName.Name,
					APIVersion: appsv1.SchemeGroupVersion.String(),
				},
				MinReplicas: ptr.New(int32(1)),
				MaxReplicas: 3,
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: v1.ResourceCPU,
							Target: autoscalingv2.MetricTarget{
								Type:               autoscalingv2.UtilizationMetricType,
								AverageUtilization: ptr.New(int32(80)),
							},
						},
					},
					{
						Type: autoscalingv2.PodsMetricSourceType,
						Pods: &autoscalingv2.PodsMetricSource{
							Metric: autoscalingv2.MetricIdentifier{
								Name: "some-metric",
							},
							Target: autoscalingv2.MetricTarget{
								Type:         autoscalingv2.AverageValueMetricType,
								AverageValue: ptr.New(resource.MustParse("10")),
							},
						},
					},
					{
						Type: autoscalingv2.ObjectMetricSourceType,
						Object: &autoscalingv2.ObjectMetricSource{
							DescribedObject: autoscalingv2.CrossVersionObjectReference{
								Kind: "Service",
								Name: "my-service",
							},
							Target: autoscalingv2.MetricTarget{
								Type:  autoscalingv2.ValueMetricType,
								Value: ptr.New(resource.MustParse("5")),
							},
							Metric: autoscalingv2.MetricIdentifier{
								Name: "object-metric",
							},
						},
					},
				},
			},
		},
	})

	s.by("Deleting the HPA")

	s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.Scale.Horizontal.CPUTarget = nil
		c.Spec.Scale.Horizontal.CustomMetrics = nil
		c.Spec.Scale.Horizontal.Instances.Max = ptr.New(uint32(1))
	})

	s.Assert().Eventually(func() bool {
		if err := s.Client.Get(ctx, nsName, &autoscalingv2.HorizontalPodAutoscaler{}); err != nil {
			if kerrors.IsNotFound(err) {
				return true
			}
		}
		return false
	}, waitFor, tick)
}

func (s *K8sTestSuite) testCronJob(ctx context.Context) {
	s.by("Creating Cron Jobs")
	s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.CronJobs = []v1alpha2.CronJob{
			{
				Name:     "job1",
				Schedule: "* * * * *",
				URL: &v1alpha2.URL{
					Port: 8000,
					Path: "/some/path",
				},
			},
			{
				Name:     "job2",
				Schedule: "10 * * * *",
				Command: &v1alpha2.JobCommand{
					Command: "./cmd",
					Args:    []string{"arg1", "arg2"},
				},
				MaxRetries:     ptr.New(uint(1)),
				TimeoutSeconds: ptr.New(uint(10)),
			},
		}
	})

	capsule, capsuleOwnerRef := s.getCapsule(ctx)
	deployment := appsv1.Deployment{}
	s.Assert().NoError(s.Client.Get(ctx, client.ObjectKeyFromObject(&capsule), &deployment))
	podTemplate := deployment.Spec.Template.DeepCopy()
	podTemplate.Spec.Containers[0].Command = []string{"./cmd"}
	podTemplate.Spec.Containers[0].Args = []string{"arg1", "arg2"}
	podTemplate.Spec.RestartPolicy = v1.RestartPolicyNever
	job1 := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job1",
			Namespace: nsName.Namespace,
			Labels: map[string]string{
				controller.LabelCapsule: nsName.Name,
				controller.LabelCron:    "job1",
			},
			OwnerReferences: []metav1.OwnerReference{capsuleOwnerRef},
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "* * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						controller.LabelCapsule: nsName.Name,
						controller.LabelCron:    "job1",
					},
					// Annotations:                map[string]string{},
				},
				Spec: batchv1.JobSpec{
					ActiveDeadlineSeconds: nil,
					BackoffLimit:          nil,
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{{
								Name:    fmt.Sprintf("%s-%s", nsName.Name, "job1"),
								Image:   "quay.io/curl/curl:latest",
								Command: []string{"curl"},
								Args:    []string{"-G", "--fail-with-body", "http://test:8000/some/path"},
							}},
							RestartPolicy: v1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}
	job2 := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job2",
			Namespace: nsName.Namespace,
			Labels: map[string]string{
				controller.LabelCapsule: nsName.Name,
				controller.LabelCron:    "job2",
			},
			OwnerReferences: []metav1.OwnerReference{capsuleOwnerRef},
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "10 * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						controller.LabelCapsule: nsName.Name,
						controller.LabelCron:    "job2",
					},
				},
				Spec: batchv1.JobSpec{
					ActiveDeadlineSeconds: ptr.New(int64(10)),
					BackoffLimit:          ptr.New(int32(1)),
					Template:              *podTemplate,
				},
			},
		},
	}
	s.expectResources(ctx, []client.Object{job1, job2})

	s.by("Deleting one Cron Job")
	s.updateCapsule(ctx, func(c *v1alpha2.Capsule) {
		c.Spec.CronJobs = capsule.Spec.CronJobs[:1]
	})
	s.expectResources(ctx, []client.Object{job1})
}

func (s *K8sTestSuite) testPrometheusServiceMonitor(ctx context.Context) {
	s.by("Creating Prometheus Service Monitor")
	_, capsuleOwnerRef := s.getCapsule(ctx)
	s.expectResources(ctx, []client.Object{
		&monitorv1.ServiceMonitor{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsName.Name,
				Namespace: nsName.Namespace,
				Labels: map[string]string{
					controller.LabelCapsule: nsName.Name,
				},
				OwnerReferences: []metav1.OwnerReference{capsuleOwnerRef},
			},
			Spec: monitorv1.ServiceMonitorSpec{
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						controller.LabelCapsule: nsName.Name,
					},
				},
				Endpoints: []monitorv1.Endpoint{{
					Port: "metricsport",
					Path: "metrics",
				}},
			},
		},
	})
}

func (s *K8sTestSuite) testDeleteCapsule(ctx context.Context) {
	capsule, _ := s.getCapsule(ctx)
	s.by("Deleting the capsule")
	s.Assert().NoError(s.Client.Delete(ctx, &capsule))
	s.Assert().Eventually(func() bool {
		if err := s.Client.Get(ctx, nsName, &capsule); err != nil {
			if kerrors.IsNotFound(err) {
				return true
			}
		}
		return false
	}, waitFor, tick)
}

func (s *K8sTestSuite) by(msg string) {
	s.T().Log("STEP: ", msg)
}
