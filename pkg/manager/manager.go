package manager

import (
	"fmt"
	"os"

	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func getEnvWithDefault(env, def string) string {
	v := os.Getenv(env)
	if v == "" {
		return def
	}
	return v
}

func New(
	cfgS config.Service,
	scheme *runtime.Scheme,
	capabilitiesService capabilities.Service,
	pluginManager *plugin.Manager,
) (manager.Manager, error) {
	cfg := cfgS.Operator()

	logger := zap.New(zap.UseDevMode(cfg.DevModeEnabled))

	restConfig := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
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

	clientSet, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	cr := &controller.CapsuleReconciler{
		Client:              mgr.GetClient(),
		Scheme:              mgr.GetScheme(),
		Config:              cfg,
		ClientSet:           clientSet,
		CapabilitiesService: capabilitiesService,
		PluginManager:       pluginManager,
	}

	if err := cr.SetupWithManager(mgr, logger); err != nil {
		return nil, err
	}

	pr := controller.NewProjectEnvironmentController(mgr.GetClient(), mgr.GetScheme(), cfg, clientSet, logger)
	if err := pr.SetupWithManager(mgr); err != nil {
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

		if err := mgr.AddHealthzCheck("webhooks", mgr.GetWebhookServer().StartedChecker()); err != nil {
			return nil, fmt.Errorf("could not add webhooks healthz check: %w", err)
		}
		if err := mgr.AddReadyzCheck("webhooks", mgr.GetWebhookServer().StartedChecker()); err != nil {
			return nil, fmt.Errorf("could not add webhooks readyz check: %w", err)
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
