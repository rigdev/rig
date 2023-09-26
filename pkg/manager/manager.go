package manager

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/rigdev/rig/internal/controller"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	rigdevv1alpha1 "github.com/rigdev/rig/pkg/api/v1alpha1"
)

func NewScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(configv1alpha1.AddToScheme(s))
	utilruntime.Must(v1alpha1.AddToScheme(s))
	return s
}

func NewManager(cfg *configv1alpha1.OperatorConfig, scheme *runtime.Scheme) (manager.Manager, error) {
	logger := zap.New(zap.UseDevMode(cfg.DevModeEnabled))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		Metrics:                 metricsserver.Options{BindAddress: ":8080"},
		HealthProbeBindAddress:  ":8081",
		Logger:                  logger,
		LeaderElection:          *cfg.LeaderElectionEnabled,
		LeaderElectionID:        "3d9f417a.rig.dev",
		LeaderElectionNamespace: "rig-system",
	})
	if err != nil {
		return nil, err
	}

	cr := &controller.CapsuleReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	if err := cr.SetupWithManager(mgr); err != nil {
		return nil, err
	}

	if *cfg.WebhooksEnabled {
		if err := (&rigdevv1alpha1.Capsule{}).SetupWebhookWithManager(mgr); err != nil {
			return nil, err
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, err
	}

	return mgr, err
}
