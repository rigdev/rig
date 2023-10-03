package k8s_test

import (
	"context"
	"fmt"
	"testing"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rigdev/rig/internal/controller"
	"github.com/rigdev/rig/pkg/api/v1alpha1"

	//+kubebuilder:scaffold:imports

	"github.com/rigdev/rig/pkg/ptr"
)

func TestIntegrationCapsuleReconcilerNginx(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	env := setupTest(t, options{runManager: true})
	defer env.cancel()
	k8sClient := env.k8sClient

	ctx := context.Background()
	nsName := types.NamespacedName{
		Name:      "test",
		Namespace: "nginx",
	}

	ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName.Namespace}}
	assert.NoError(t, k8sClient.Create(ctx, ns))

	capsule := v1alpha1.Capsule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsName.Name,
			Namespace: nsName.Namespace,
		},
		Spec: v1alpha1.CapsuleSpec{
			Image: "nginx:1.25.1",
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

	err := k8sClient.Get(ctx, nsName, &v1.Service{})
	assert.True(t, kerrors.IsNotFound(err))

	capsule.Spec.Interfaces = []v1alpha1.CapsuleInterface{
		{
			Name: "http",
			Port: 80,
		},
	}

	assert.NoError(t, k8sClient.Update(ctx, &capsule))

	var svc v1.Service
	assert.Eventually(t, func() bool {
		if err := k8sClient.Get(ctx, nsName, &svc); err != nil {
			t.Logf("could not get svc: %s", err.Error())
			return false
		}
		return true
	}, waitFor, tick)

	if assert.Len(t, svc.Spec.Ports, 1) {
		assert.Equal(t, capsule.Spec.Interfaces[0].Name, svc.Spec.Ports[0].Name)
		assert.Equal(t, capsule.Spec.Interfaces[0].Port, svc.Spec.Ports[0].Port)
		assert.Equal(t, capsule.Spec.Interfaces[0].Name, svc.Spec.Ports[0].TargetPort.StrVal)
	}
	if assert.Len(t, svc.OwnerReferences, 1) {
		assert.Equal(t, capsuleOwnerRef, svc.OwnerReferences[0])
	}

	err = k8sClient.Get(ctx, nsName, &netv1.Ingress{})
	assert.True(t, kerrors.IsNotFound(err))

	capsule.Spec.Interfaces[0].Public = &v1alpha1.CapsulePublicInterface{
		Ingress: &v1alpha1.CapsuleInterfaceIngress{
			Host: "test.com",
		},
	}

	assert.NoError(t, k8sClient.Update(ctx, &capsule))

	var ing netv1.Ingress
	assert.Eventually(t, func() bool {
		if err := k8sClient.Get(ctx, nsName, &ing); err != nil {
			return false
		}
		return true
	}, waitFor, tick)

	if assert.Len(t, ing.Spec.Rules, 1) {
		rule := ing.Spec.Rules[0]
		assert.Equal(t, capsule.Spec.Interfaces[0].Public.Ingress.Host, rule.Host)
		if assert.NotNil(t, rule.IngressRuleValue.HTTP) &&
			assert.Len(t, rule.IngressRuleValue.HTTP.Paths, 1) {
			path := rule.IngressRuleValue.HTTP.Paths[0]
			assert.Equal(t, ptr.New(netv1.PathTypePrefix), path.PathType)
			assert.Equal(t, "/", path.Path)
			assert.Equal(t, capsule.Name, path.Backend.Service.Name)
			assert.Equal(t, capsule.Spec.Interfaces[0].Name, path.Backend.Service.Port.Name)
		}
	}

	if assert.Len(t, ing.Spec.TLS, 1) {
		tls := ing.Spec.TLS[0]
		assert.Equal(t, fmt.Sprintf("%s-tls", capsule.Name), tls.SecretName)
		if assert.Len(t, tls.Hosts, 1) {
			assert.Equal(t, capsule.Spec.Interfaces[0].Public.Ingress.Host, tls.Hosts[0])
		}
	}

	assert.True(t, controller.IsOwnedBy(&capsule, &ing))

	var crt cmv1.Certificate
	assert.Eventually(t, func() bool {
		if err := k8sClient.Get(ctx, nsName, &crt); err != nil {
			return false
		}
		return true
	}, waitFor, tick)

	assert.True(t, controller.IsOwnedBy(&capsule, &crt))
	assert.Equal(t, fmt.Sprintf("%s-tls", capsule.Name), crt.Spec.SecretName)
	assert.Equal(t, cmv1.ClusterIssuerKind, crt.Spec.IssuerRef.Kind)
	assert.Equal(t, "test", crt.Spec.IssuerRef.Name)

	if assert.Len(t, crt.Spec.DNSNames, 1) {
		assert.Equal(t, capsule.Spec.Interfaces[0].Public.Ingress.Host, crt.Spec.DNSNames[0])
	}

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

	var lb v1.Service
	assert.Eventually(t, func() bool {
		if err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      fmt.Sprintf("%s-lb", nsName.Name),
			Namespace: nsName.Namespace,
		}, &lb); err != nil {
			return false
		}
		return true
	}, waitFor, tick)

	assert.Equal(t, v1.ServiceTypeLoadBalancer, lb.Spec.Type)
	if assert.Len(t, lb.Spec.Ports, 1) {
		p := lb.Spec.Ports[0]
		assert.Equal(t, "http", p.Name)
		assert.Equal(t, int32(1), p.Port)
		assert.Equal(t, "http", p.TargetPort.StrVal)
	}
	if assert.Len(t, lb.OwnerReferences, 1) {
		assert.Equal(t, capsuleOwnerRef, lb.OwnerReferences[0])
	}

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
