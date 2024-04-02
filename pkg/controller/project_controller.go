package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ProjectController reconciles a Project object
type ProjectController struct {
	client.Client
	Scheme    *runtime.Scheme
	Config    *configv1alpha1.OperatorConfig
	ClientSet clientset.Interface
	Logger    logr.Logger
}

func NewProjectController(
	c client.Client,
	scheme *runtime.Scheme,
	config *configv1alpha1.OperatorConfig,
	clientSet clientset.Interface,
	logger logr.Logger,
) *ProjectController {
	logger = logger.WithValues("crd", "project")
	return &ProjectController{
		Client:    c,
		Scheme:    scheme,
		Config:    config,
		ClientSet: clientSet,
		Logger:    logger,
	}
}

func (p *ProjectController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.Project{}).
		Owns(&corev1.Namespace{}).
		Complete(p)
}

//+kubebuilder:rbac:groups=rig.dev,resources=projects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rig.dev,resources=projects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="core",resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (p *ProjectController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	project := &v1alpha2.Project{}
	if err := p.Get(ctx, req.NamespacedName, project); err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("could not fetch Project: %w", err)
	}

	request := pipeline.NewProjectRequest(p.Client, p, p.Config, p.Scheme, p.Logger, project)

	if _, err := pipeline.ExecuteRequest(ctx, request, projectSteps, true); err != nil {
		p.Logger.Error(err, "reconciliation ended with error")
		return ctrl.Result{}, err
	}

	p.Logger.Info("reconciliation completed successfully")

	return ctrl.Result{}, nil
}

var projectSteps = []pipeline.Step[pipeline.ProjectRequest]{
	namespaceStep{},
}

type namespaceStep struct{}

func (n namespaceStep) Apply(_ context.Context, req pipeline.ProjectRequest) error {
	project := req.Project()

	for _, ns := range project.Spec.Namespaces {
		if err := req.Set(&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
