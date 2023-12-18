package k8s_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-logr/logr/testr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/manager"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	//+kubebuilder:scaffold:imports
)

const (
	waitFor = time.Second * 10
	tick    = time.Millisecond * 200
)

type K8sTestSuite struct {
	suite.Suite

	cancel  context.CancelFunc
	TestEnv *envtest.Environment
	Client  client.Client
}

func (s *K8sTestSuite) SetupSuite() {
	setupDone := false
	defer func() {
		if !setupDone {
			s.TearDownSuite()
		}
	}()

	t := s.Suite.T()

	s.TestEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "deploy", "kustomize", "crd", "bases"),
			filepath.Join("."),
		},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: filepath.Join("..", "..", "..", "tools", "bin", "k8s",
			fmt.Sprintf("1.28.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	var err error
	cfg, err := s.TestEnv.Start()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	scheme := manager.NewScheme()
	manager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme,
		Metrics: server.Options{BindAddress: "0"},
		Logger:  testr.New(t),
	})
	require.NoError(t, err)

	clientSet, err := clientset.NewForConfig(cfg)
	require.NoError(t, err)

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	require.NoError(t, err)
	require.NotNil(t, k8sClient)
	s.Client = k8sClient

	capsuleReconciler := &controller.CapsuleReconciler{
		Client: manager.GetClient(),
		Scheme: scheme,
		Config: &configv1alpha1.OperatorConfig{
			Certmanager: &configv1alpha1.CertManagerConfig{
				ClusterIssuer:              "test",
				CreateCertificateResources: true,
			},
			Ingress: configv1alpha1.IngressConfig{
				PathType: netv1.PathTypeExact,
			},
		},
		ClientSet: clientSet,
	}

	require.NoError(t, capsuleReconciler.SetupWithManager(manager))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		require.NoError(t, manager.Start(ctx))
	}()

	s.cancel = cancel
	setupDone = true
}

func (s *K8sTestSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.TestEnv != nil {
		if err := s.TestEnv.Stop(); err != nil {
			fmt.Println(err)
		}
	}
}

func TestIntegrationK8s(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	suite.Run(t, &K8sTestSuite{})
}
