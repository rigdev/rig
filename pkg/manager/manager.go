package manager

import (
	"os"

	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/service/config"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func NewScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientsetscheme.AddToScheme(s))
	utilruntime.Must(configv1alpha1.AddToScheme(s))
	utilruntime.Must(v1alpha1.AddToScheme(s))
	utilruntime.Must(v1alpha2.AddToScheme(s))
	utilruntime.Must(certv1.AddToScheme(s))
	utilruntime.Must(monitorv1.AddToScheme(s))
	return s
}

func getEnvWithDefault(env, def string) string {
	v := os.Getenv(env)
	if v == "" {
		return def
	}
	return v
}

func NewManager(cfgS config.Service, scheme *runtime.Scheme) (manager.Manager, error) {
	cfg := cfgS.Get()

	logger := zap.New(zap.UseDevMode(cfg.DevModeEnabled))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                        scheme,
		Metrics:                       metricsserver.Options{BindAddress: ":8080"},
		HealthProbeBindAddress:        ":8081",
		Logger:                        logger,
		LeaderElection:                *cfg.LeaderElectionEnabled,
		LeaderElectionID:              "3d9f417a.rig.dev",
		LeaderElectionNamespace:       getEnvWithDefault("POD_NAMESPACE", "rig-system"),
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		return nil, err
	}

	cr := &controller.CapsuleReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: cfg,
	}

	if err := cr.SetupWithManager(mgr); err != nil {
		return nil, err
	}

	if *cfg.WebhooksEnabled {
		if err := (&v1alpha1.Capsule{}).SetupWebhookWithManager(mgr); err != nil {
			return nil, err
		}
		if err := (&v1alpha2.Capsule{}).SetupWebhookWithManager(mgr); err != nil {
			return nil, err
		}
		//+kubebuilder:scaffold:builder
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, err
	}

	return mgr, err
}
