package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var capsulelog = logf.Log.WithName("capsule-resource")

func (r *Capsule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-rig-dev-v1alpha1-capsule,mutating=true,failurePolicy=fail,sideEffects=None,groups=rig.dev,resources=capsules,verbs=create;update,versions=v1alpha1,name=mcapsule.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Capsule{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Capsule) Default() {
	capsulelog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-rig-dev-v1alpha1-capsule,mutating=false,failurePolicy=fail,sideEffects=None,groups=rig.dev,resources=capsules,verbs=create;update,versions=v1alpha1,name=vcapsule.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Capsule{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Capsule) ValidateCreate() (admission.Warnings, error) {
	capsulelog.Info("validate create", "name", r.Name)
	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Capsule) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	capsulelog.Info("validate update", "name", r.Name)
	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Capsule) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (r *Capsule) validate() (admission.Warnings, error) {
	warns, errs := r.validateInterfaces()
	return warns, errs.ToAggregate()
}

func (r *Capsule) validateInterfaces() (admission.Warnings, field.ErrorList) {
	if len(r.Spec.Interfaces) == 0 {
		return nil, nil
	}

	var errs field.ErrorList

	names := map[string]struct{}{}
	ports := map[int32]struct{}{}
	infsPath := field.NewPath("spec").Child("interfaces")
	for i, inf := range r.Spec.Interfaces {
		infPath := infsPath.Index(i)

		if _, ok := names[inf.Name]; ok {
			errs = append(errs, field.Duplicate(infPath.Child("name"), inf.Name))
		} else {
			names[inf.Name] = struct{}{}
		}

		if _, ok := ports[inf.Port]; ok {
			errs = append(errs, field.Duplicate(infPath.Child("port"), inf.Port))
		} else {
			ports[inf.Port] = struct{}{}
		}

		if inf.Public != nil {
			public := inf.Public
			publicPath := infPath.Child("public")
			if public.Ingress == nil && public.LoadBalancer == nil {
				errs = append(errs, field.Required(publicPath, "ingress or loadBalancer is required"))
			}
			if public.Ingress != nil && public.LoadBalancer != nil {
				errs = append(errs, field.Invalid(publicPath, public, "ingress and loadBalancer are mutually exclusive"))
			}
		}
	}

	return nil, errs
}

func (r *Capsule) validateFiles() (admission.Warnings, field.ErrorList) {
	// TODO
	return nil, nil
}
