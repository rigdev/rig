package v1alpha1

import (
	"github.com/rigdev/rig/pkg/ptr"
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
	if r.Spec.Replicas == nil {
		r.Spec.Replicas = ptr.New(int32(1))
	}
	if r.Spec.HorizontalScale.MinReplicas == nil {
		r.Spec.HorizontalScale.MinReplicas = ptr.New(uint32(1))
	}
	if r.Spec.HorizontalScale.MaxReplicas == nil {
		max := *r.Spec.HorizontalScale.MinReplicas
		r.Spec.HorizontalScale.MaxReplicas = ptr.New(max)
	}
	if r.Spec.Env != nil && r.Spec.Env.Automatic == nil {
		r.Spec.Env.Automatic = ptr.New(true)
	}
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
	var (
		allWarns admission.Warnings
		allErrs  field.ErrorList
		warns    admission.Warnings
		errs     field.ErrorList
	)

	warns, errs = r.validateSpec()
	allWarns = append(allWarns, warns...)
	allErrs = append(allErrs, errs...)

	warns, errs = r.validateInterfaces()
	allWarns = append(allWarns, warns...)
	allErrs = append(allErrs, errs...)

	warns, errs = r.validateFiles()
	allWarns = append(allWarns, warns...)
	allErrs = append(allErrs, errs...)

	errs = append(errs, r.Spec.HorizontalScale.validate(field.NewPath("horizontalScale"))...)

	return allWarns, allErrs.ToAggregate()
}

func (r *Capsule) validateSpec() (admission.Warnings, field.ErrorList) {
	var errs field.ErrorList

	if r.Spec.Image == "" {
		errs = append(errs, field.Required(field.NewPath("spec").Child("image"), ""))
	}
	return nil, errs
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

		if inf.Name == "" {
			errs = append(errs, field.Required(infPath.Child("name"), ""))
		}

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
			if public.Ingress != nil && public.Ingress.Host == "" {
				errs = append(errs, field.Required(publicPath.Child("ingress").Child("host"), ""))
			}
			if public.LoadBalancer != nil {
				p := public.LoadBalancer.NodePort
				if p > 0 && (p < 30000 || p > 32767) {
					errs = append(errs, field.Invalid(publicPath.Child("loadBalancer").Child("nodePort"), p, "nodePort must be in the range [30,000; 32,767]"))
				}
			}
		}
	}

	return nil, errs
}

func (r *Capsule) validateFiles() (admission.Warnings, field.ErrorList) {
	var errs field.ErrorList

	paths := map[string]struct{}{}
	filesPath := field.NewPath("spec").Child("files")
	for i, f := range r.Spec.Files {
		fPath := filesPath.Index(i)

		if f.Path == "" {
			errs = append(errs, field.Required(fPath.Child("path"), ""))
		}

		if _, ok := paths[f.Path]; ok {
			errs = append(errs, field.Duplicate(fPath.Child("path"), f.Path))
		} else {
			paths[f.Path] = struct{}{}
		}

		if f.Secret != nil && f.ConfigMap != nil {
			errs = append(errs, field.Invalid(fPath, f, "configMap and secret are mutually exclusive"))
		}
		if f.Secret == nil && f.ConfigMap == nil {
			errs = append(errs, field.Required(fPath, "one of configMap or secret is required"))
		}

		if f.Secret != nil {
			if f.Secret.Name == "" {
				errs = append(errs, field.Required(fPath.Child("secret").Child("name"), ""))
			}
			if f.Secret.Key == "" {
				errs = append(errs, field.Required(fPath.Child("secret").Child("key"), ""))
			}
		}

		if f.ConfigMap != nil {
			if f.ConfigMap.Name == "" {
				errs = append(errs, field.Required(fPath.Child("configMap").Child("name"), ""))
			}
			if f.ConfigMap.Key == "" {
				errs = append(errs, field.Required(fPath.Child("configMap").Child("key"), ""))
			}
		}
	}

	return nil, errs
}

func (h *HorizontalScale) validate(fPath *field.Path) field.ErrorList {
	if h == nil {
		return nil
	}

	var errs field.ErrorList

	var maxReplicas uint32
	var minReplicas uint32
	if h.MinReplicas == nil {
		minReplicas = 1
	} else {
		minReplicas = *h.MinReplicas
	}
	if h.MaxReplicas == nil {
		maxReplicas = minReplicas
	} else {
		maxReplicas = *h.MaxReplicas
	}

	if maxReplicas > 0 && maxReplicas < minReplicas {
		errs = append(errs, field.Invalid(fPath.Child("maxReplicas"), maxReplicas, "maxReplicas cannot be smaller than minReplicas"))
	}

	avg := h.CPUTarget.AverageUtilizationPercentage
	if avg > 100 {
		errs = append(errs, field.Invalid(fPath.Child("cpuTarget").Child("averageUtilizationPercentage"), avg, "cannot be larger than 100"))
	}

	return errs
}
