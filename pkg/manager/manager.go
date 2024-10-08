package manager

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	cfg_v1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller"
	"github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/objectstatus"
	"github.com/rigdev/rig/pkg/service/pipeline"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
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
	cfg *cfg_v1alpha1.OperatorConfig,
	scheme *runtime.Scheme,
	capabilitiesService capabilities.Service,
	pipeline pipeline.Service,
	objectstatus objectstatus.Service,
	restConfig *rest.Config,
	logger logr.Logger,
	lc fx.Lifecycle,
) (manager.Manager, error) {
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

	cr := &controller.CapsuleReconciler{
		Client:              mgr.GetClient(),
		Scheme:              scheme,
		Config:              cfg,
		CapabilitiesService: capabilitiesService,
		PipelineService:     pipeline,
		ObjectStatusService: objectstatus,
		Lifecycle:           lc,
	}

	if err := cr.SetupWithManager(mgr, ""); err != nil {
		return nil, err
	}

	// pr := controller.NewProjectEnvironmentController(mgr.GetClient(), mgr.GetScheme(), cfg, clientSet, logger)
	// if err := pr.SetupWithManager(mgr); err != nil {
	// 	return nil, err
	// }

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
