package k8s_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/go-logr/logr/testr"
	"github.com/nsf/jsondiff"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/objectstatus"
	"github.com/rigdev/rig/pkg/service/pipeline"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx/fxtest"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
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
	Suite

	cancel  context.CancelFunc
	TestEnv *envtest.Environment
}

func (s *K8sTestSuite) SetupSuite() {
	setupDone := false
	defer func() {
		if !setupDone {
			s.TearDownSuite()
		}
	}()

	t := s.Suite.T()

	scheme := scheme.New()
	require.NoError(t, cmv1.AddToScheme(scheme))

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
			RoutesStep: configv1alpha1.CapsuleStep{
				Plugin: "rigdev.ingress_routes",
				Config: `
clusterIssuer: "test"
createCertificateResources: true
ingressClassName: "nginx"
disableTLS: false
`,
			},
			ServiceMonitorStep: configv1alpha1.CapsuleStep{
				Plugin: "rigdev.service_monitor",
				Config: `
path: "metrics"
portName: "metricsport"`,
			},
		},
	}

	cc, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	require.NoError(t, err)

	wd, err := os.Getwd()
	require.NoError(t, err)
	builtinBinPath := path.Join(path.Dir(path.Dir(path.Dir(wd))), "bin", "rig-operator")
	pmanager, err := plugin.NewManager(cfg, plugin.SetBuiltinBinaryPathOption(builtinBinPath))
	require.NoError(t, err)

	cs := capabilities.NewService(cc, clientSet.Discovery(), nil)

	lc := fxtest.NewLifecycle(t)
	ps := pipeline.NewService(opConfig, cc, cs, ctrl.Log, pmanager, lc, nil)

	capsuleReconciler := &controller.CapsuleReconciler{
		Client:              manager.GetClient(),
		Scheme:              scheme,
		Config:              opConfig,
		CapabilitiesService: cs,
		PipelineService:     ps,
		ObjectStatusService: objectstatus.NewService(opConfig, ps, ctrl.Log),
		Lifecycle:           lc,
	}

	require.NoError(t, capsuleReconciler.SetupWithManager(manager, "capsule"))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		require.NoError(t, manager.Start(ctx))
	}()

	require.NoError(t, lc.Start(context.Background()))

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

type Suite struct {
	suite.Suite
	Client client.Client
}

func (s *Suite) expectResources(ctx context.Context, resources []client.Object) {
	for _, expectedResource := range resources {
		count := 0
		currentResource := expectedResource.DeepCopyObject().(client.Object)
		for {
			if err := s.Client.Get(ctx, client.ObjectKeyFromObject(expectedResource), currentResource); kerrors.IsNotFound(err) {
				time.Sleep(100 * time.Millisecond)
				continue
			} else if err != nil {
				s.Require().NoError(err)
			}

			// Clear this property.
			currentResource.SetCreationTimestamp(metav1.Time{})

			expectedBytes, err := json.Marshal(expectedResource)
			s.Require().NoError(err)

			currentBytes, err := json.Marshal(currentResource)
			s.Require().NoError(err)

			opt := jsondiff.DefaultConsoleOptions()
			diff, change := jsondiff.Compare(currentBytes, expectedBytes, &opt)

			count++
			if jsondiff.SupersetMatch == diff {
				break
			} else if count > 20 {
				s.Require().Equal(jsondiff.SupersetMatch, diff, change)
			}

			time.Sleep(250 * time.Millisecond)
		}
	}
}
