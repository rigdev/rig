package operator

import (
	"context"

	"go.uber.org/fx"
	zapcore "go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type Service interface{}

type service struct {
	log    *zapcore.Logger
	cancel context.CancelFunc
}

type NewParams struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger *zapcore.Logger
}

func New(p NewParams) Service {
	s := &service{
		log: p.Logger,
	}

	p.Lifecycle.Append(fx.StartStopHook(s.start, s.stop))

	return s
}

func (s *service) start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// TODO use the logger from s.log but add the KubeAwareEncoder from sigs.k8s.io/controller-runtime/pkg/log/zap
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: ":8080"},
		HealthProbeBindAddress: ":8081",
		LeaderElection:         true,
		LeaderElectionID:       "3d9f417a.rig.dev",
	})
	if err != nil {
		s.log.Error("unable to start manager", zapcore.Error(err))
		return
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		s.log.Error("unable to set up health check", zapcore.Error(err))
		return
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		s.log.Error("unable to set up ready check", zapcore.Error(err))
		return
	}

	go func() {
		s.log.Info("starting operator service")
		if err := mgr.Start(ctx); err != nil {
			s.log.Error("problem running manager", zapcore.Error(err))
			return
		}
	}()
}

func (s *service) stop() {
	s.cancel()
}
