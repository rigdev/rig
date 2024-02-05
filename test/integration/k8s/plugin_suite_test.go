package k8s_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-logr/logr/testr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
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

type PluginTestSuite struct {
	Suite

	cancel  context.CancelFunc
	TestEnv *envtest.Environment
}

func (s *PluginTestSuite) SetupSuite() {
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

	// Enable sidecars.
	s.TestEnv.ControlPlane.GetAPIServer().Configure().Append("feature-gates", "SidecarContainers=true")

	var err error
	cfg, err := s.TestEnv.Start()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	scheme := scheme.New()
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

	opConfig := &configv1alpha1.OperatorConfig{
		Certmanager: &configv1alpha1.CertManagerConfig{
			ClusterIssuer:              "test",
			CreateCertificateResources: true,
		},
		Ingress: configv1alpha1.IngressConfig{
			PathType: netv1.PathTypeExact,
		},
		PrometheusServiceMonitor: &configv1alpha1.PrometheusServiceMonitor{
			Path:     "metrics",
			PortName: "metricsport",
		},
		Steps: []configv1alpha1.Step{
			{
				Plugin: configv1alpha1.Plugin{
					Object: &configv1alpha1.ObjectPlugin{
						Group: "apps",
						Kind:  "Deployment",
						Object: `
spec:
  replicas: 2`,
					},
				},
			},
			{
				Plugin: configv1alpha1.Plugin{
					Sidecar: &configv1alpha1.SidecarPlugin{
						Container: `
image: nginx
name: nginx`,
					},
				},
			},
			{
				Plugin: configv1alpha1.Plugin{
					InitContainer: &configv1alpha1.InitContainerPlugin{
						Container: `
image: alpine
name: startup
command: ['sh', '-c', 'echo Hello']`,
					},
				},
			},
		},
	}

	configService := config.NewServiceFromConfigs(opConfig, nil)

	cc, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	require.NoError(t, err)

	capsuleReconciler := &controller.CapsuleReconciler{
		Client:              manager.GetClient(),
		Scheme:              scheme,
		Config:              opConfig,
		ClientSet:           clientSet,
		CapabilitiesService: capabilities.NewService(configService, cc, clientSet.Discovery()),
	}

	require.NoError(t, capsuleReconciler.SetupWithManager(manager))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		require.NoError(t, manager.Start(ctx))
	}()

	s.cancel = cancel
	setupDone = true
}

func (s *PluginTestSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.TestEnv != nil {
		if err := s.TestEnv.Stop(); err != nil {
			fmt.Println(err)
		}
	}
}

func TestIntegrationPlugin(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	suite.Run(t, &PluginTestSuite{})
}
