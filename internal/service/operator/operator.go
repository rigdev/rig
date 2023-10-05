package operator

import (
	"context"

	"github.com/go-logr/zapr"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/controller"
	rigdevv1alpha1 "github.com/rigdev/rig/pkg/api/v1alpha1"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type Service interface{}

type service struct {
	log    *zap.Logger
	cfg    config.Config
	cancel context.CancelFunc
}

type NewParams struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger *zap.Logger
	Config config.Config
}

func New(p NewParams) Service {
	s := &service{
		log: p.Logger,
		cfg: p.Config,
	}

	p.Lifecycle.Append(fx.StartStopHook(s.start, s.stop))

	return s
}

func (s *service) start() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	ctrl.SetLogger(zapr.NewLogger(s.log.Named("operator")))

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(rigdevv1alpha1.AddToScheme(scheme))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		Metrics:                 metricsserver.Options{BindAddress: ":8080"},
		HealthProbeBindAddress:  ":8081",
		LeaderElection:          true,
		LeaderElectionID:        "3d9f417a.rig.dev",
		LeaderElectionNamespace: "rig-system",
	})
	if err != nil {
		s.log.Error("unable to start manager", zap.Error(err))
		return err
	}

	//+kubebuilder:scaffold:builder
	cr := &controller.CapsuleReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	if err := cr.SetupWithManager(mgr); err != nil {
		s.log.Error("unable to setup controller", zap.Error(err))
		return err
	}

	if s.cfg.Client.Kubernetes.WebhooksEnabled {
		if err := (&rigdevv1alpha1.Capsule{}).SetupWebhookWithManager(mgr); err != nil {
			s.log.Error("could not setup webhook with manager", zap.Error(err))
			return err
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		s.log.Error("unable to set up health check", zap.Error(err))
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		s.log.Error("unable to set up ready check", zap.Error(err))
		return err
	}

	go func() {
		s.log.Info("starting operator service")
		if err := mgr.Start(ctx); err != nil {
			s.log.Fatal("problem running manager", zap.Error(err))
			return
		}
	}()

	return nil
}

func (s *service) stop() {
	s.cancel()
}
