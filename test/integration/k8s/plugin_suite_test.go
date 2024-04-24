package k8s_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-logr/logr/testr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/pipeline"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx/fxtest"
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

	scheme := scheme.New()
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
		Pipeline: configv1alpha1.Pipeline{
			ServiceMonitorStep: configv1alpha1.CapsuleStep{
				Plugin: "rigdev.service_monitor",
				Config: `
path: "metrics"
portName: "metricsport"`,
			},
			Steps: []configv1alpha1.Step{
				{
					Plugins: []configv1alpha1.Plugin{
						{
							Name: "rigdev.object_template",
							Config: `
group: "apps"
kind: "Deployment"
object: |
  spec:
    replicas: 2
`,
						},
						{
							Name: "rigdev.sidecar",
							Config: `
container:
  image: nginx
  name: nginx
`,
						},
						{
							Name: "rigdev.init_container",
							Config: `
container:
  image: alpine
  name: startup
  command: ['sh', '-c', 'echo Hello']
`,
						},
					},
				},
			},
		},
	}

	cc, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	require.NoError(t, err)

	cs := capabilities.NewService(cc, clientSet.Discovery(), nil)

	wd, err := os.Getwd()
	require.NoError(t, err)
	builtinBinPath := path.Join(path.Dir(path.Dir(path.Dir(wd))), "bin", "rig-operator")
	pmanager, err := plugin.NewManager(cfg, plugin.SetBuiltinBinaryPathOption(builtinBinPath))
	require.NoError(t, err)
	lc := fxtest.NewLifecycle(t)
	ps := pipeline.NewService(opConfig, cc, cs, ctrl.Log, pmanager, lc)
	require.NoError(t, lc.Start(context.Background()))
	capsuleReconciler := &controller.CapsuleReconciler{
		Client:              manager.GetClient(),
		Scheme:              scheme,
		Config:              opConfig,
		CapabilitiesService: cs,
		PipelineService:     ps,
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
