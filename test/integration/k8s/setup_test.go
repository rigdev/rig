package k8s_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/rigdev/rig/pkg/controller"

	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/manager"
	//+kubebuilder:scaffold:imports
)

const (
	waitFor = time.Second * 10
	tick    = time.Millisecond * 200
)

type options struct {
	runManager bool
}

type env struct {
	k8sClient client.Client
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func (e *env) stop() {
	e.cancel()
	e.wg.Wait()
}

func setupTest(t *testing.T, opts options) *env {
	logf.SetLogger(testr.New(t))

	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "deploy", "kustomize", "crd", "bases"),
			filepath.Join("."),
		},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: filepath.Join("..", "..", "..", "tools", "bin", "k8s",
			fmt.Sprintf("1.28.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	var err error
	cfg, err := testEnv.Start()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	scheme := manager.NewScheme()

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	assert.NoError(t, err)
	assert.NotNil(t, k8sClient)

	ctx, cancel := context.WithCancel(context.Background())

	e := &env{
		k8sClient: k8sClient,
		cancel:    cancel,
		wg:        sync.WaitGroup{},
	}

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		<-ctx.Done()
		if err := testEnv.Stop(); err != nil {
			fmt.Printf("could not stop envtest: %s\n", err)
		}
	}()

	if opts.runManager {
		manager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  scheme,
			Metrics: server.Options{BindAddress: "0"},
		})
		assert.NoError(t, err)

		capsuleReconciler := &controller.CapsuleReconciler{
			Client: k8sClient,
			Scheme: scheme,
			Config: &configv1alpha1.OperatorConfig{
				Certmanager: &configv1alpha1.CertManagerConfig{
					ClusterIssuer:              "test",
					CreateCertificateResources: true,
				},
			},
		}

		assert.NoError(t, capsuleReconciler.SetupWithManager(manager))

		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
			assert.NoError(t, manager.Start(ctx))
		}()
	}

	return e
}
