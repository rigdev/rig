package controller_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/rigdev/rig/internal/controller"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	rigdevv1alpha1 "github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/ptr"
	//+kubebuilder:scaffold:imports
)

const (
	waitFor = time.Second * 10
	tick    = time.Millisecond * 200
)

var (
	cfg           *rest.Config
	k8sClient     client.Client
	testEnv       *envtest.Environment
	managerCancel context.CancelFunc
)

func setupTest(t *testing.T) {
	logf.SetLogger(testr.New(t))
	// TODO: find a way to use the improved implementation from controller-runtime zap
	//logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "..", "deploy", "kustomize", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "..", "..", "tools", "bin", "k8s",
			fmt.Sprintf("1.28.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	var err error
	cfg, err = testEnv.Start()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	err = rigdevv1alpha1.AddToScheme(scheme.Scheme)
	assert.NoError(t, err)

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	assert.NoError(t, err)
	assert.NotNil(t, k8sClient)

	manager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: server.Options{BindAddress: "0"},
	})
	assert.NoError(t, err)

	capsuleReconciler := &controller.CapsuleReconciler{
		Client: k8sClient,
		Scheme: manager.GetScheme(),
	}

	assert.NoError(t, capsuleReconciler.SetupWithManager(manager))

	var managerCtx context.Context
	managerCtx, managerCancel = context.WithCancel(context.Background())
	go func() {
		assert.NoError(t, manager.Start(managerCtx))
	}()
}

func tearDownTest(t *testing.T) {
	if managerCancel != nil {
		managerCancel()
	}
	if testEnv != nil {
		err := testEnv.Stop()
		assert.NoError(t, err)
	}
}

func TestIntegrationCapsuleReconcilerNginx(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	setupTest(t)
	defer tearDownTest(t)

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
