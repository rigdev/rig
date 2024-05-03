package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/uuid"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ProjectEnvironmentController reconciles a Project object
type ProjectEnvironmentController struct {
	client.Client
	Scheme    *runtime.Scheme
	Config    *configv1alpha1.OperatorConfig
	ClientSet clientset.Interface
	Logger    logr.Logger
}

func NewProjectEnvironmentController(
	c client.Client,
	scheme *runtime.Scheme,
	config *configv1alpha1.OperatorConfig,
	clientSet clientset.Interface,
	logger logr.Logger,
) *ProjectEnvironmentController {
	logger = logger.WithValues("crd", "projectenvironment")
	return &ProjectEnvironmentController{
		Client:    c,
		Scheme:    scheme,
		Config:    config,
		ClientSet: clientSet,
		Logger:    logger,
	}
}

func (p *ProjectEnvironmentController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.ProjectEnvironment{}).
		Owns(&corev1.Namespace{}).
		Complete(p)
}

//+kubebuilder:rbac:groups=rig.dev,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rig.dev,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="core",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (p *ProjectEnvironmentController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	projectEnv := &v1alpha2.ProjectEnvironment{}
	if err := p.Get(ctx, req.NamespacedName, projectEnv); err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("could not fetch ProjectEnvironment: %w", err)
	}

	request := pipeline.NewProjectEnvironmentRequest(p.Client, p, p.Config, p.Scheme, p.Logger, projectEnv)

	if _, err := pipeline.ExecuteRequest(ctx, request, projectSteps, true); err != nil {
		p.Logger.Error(err, "reconciliation ended with error")
		return ctrl.Result{}, err
	}

	p.Logger.Info("reconciliation completed successfully")

	return ctrl.Result{}, nil
}

var projectSteps = []pipeline.Step[pipeline.ProjectEnvironmentRequest]{
	namespaceStep{},
}

type namespaceStep struct{}

func (s namespaceStep) Apply(_ context.Context, req pipeline.ProjectEnvironmentRequest) error {
	projectEnv := req.ProjectEnvironment()
	if err := req.Set(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectEnv.Name,
		},
	}); err != nil {
		return err
	}

	return nil
}

func (s namespaceStep) WatchObjectStatus(
	_ context.Context,
	_ string,
	_ string,
	_ pipeline.ObjectStatusCallback,
) error {
	return nil
}

func (s namespaceStep) PluginIDs() []uuid.UUID {
	return nil
}
