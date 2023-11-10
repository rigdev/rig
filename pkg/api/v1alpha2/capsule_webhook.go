package v1alpha2

import (
	"path"

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

//+kubebuilder:webhook:path=/mutate-rig-dev-v1alpha2-capsule,mutating=true,failurePolicy=fail,sideEffects=None,groups=rig.dev,resources=capsules,verbs=create;update,versions=v1alpha2,name=mcapsule.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Capsule{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Capsule) Default() {
	capsulelog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-rig-dev-v1alpha2-capsule,mutating=false,failurePolicy=fail,sideEffects=None,groups=rig.dev,resources=capsules,verbs=create;update,versions=v1alpha2,name=vcapsule.kb.io,admissionReviewVersions=v1

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

	warns, errs = r.validateEnv()
	allWarns = append(allWarns, warns...)
	allErrs = append(allErrs, errs...)

	warns, errs = r.validateFiles()
	allWarns = append(allWarns, warns...)
	allErrs = append(allErrs, errs...)

	errs = append(errs, r.Spec.Scale.Horizontal.validate(field.NewPath("scale").Child("horizontal"))...)

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

	hasLiveness := false
	hasReadiness := false

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
		}

		if inf.Liveness != nil {
			if hasLiveness {
				errs = append(errs, field.Duplicate(infPath.Child("liveness"), inf.Liveness))
			}

			errs = append(errs, inf.Liveness.validate(infPath.Child("liveness"))...)

			hasLiveness = true
		}

		if inf.Readiness != nil {
			if hasReadiness {
				errs = append(errs, field.Duplicate(infPath.Child("readiness"), inf.Readiness))
			}

			errs = append(errs, inf.Readiness.validate(infPath.Child("readiness"))...)

			hasReadiness = true
		}
	}

	return nil, errs
}

func (p *InterfaceProbe) validate(pPath *field.Path) field.ErrorList {
	var errs field.ErrorList

	c := 0
	if p.Path != "" {
		if !path.IsAbs(p.Path) {
			errs = append(errs, field.Invalid(pPath.Child("path"), p.Path, "path must be an absolute path"))
		}
		c++
	}
	if p.GRPC != nil {
		c++
	}
	if p.TCP {
		c++
	}
	if c == 0 {
		errs = append(errs, field.Invalid(pPath, p, "interface probes must contain one of `path`, `tcp` or `grpc`"))
	}
	if c > 1 {
		errs = append(errs, field.Invalid(pPath, p, "interface probes must contain only one of `path`, `tcp` or `grpc`"))
	}
	return errs
}

func (r *Capsule) validateEnv() (admission.Warnings, field.ErrorList) {
	if r.Spec.Env == nil {
		return nil, nil
	}

	var errs field.ErrorList

	fromPath := field.NewPath("spec").Child("env").Child("from")
	for i, r := range r.Spec.Env.From {
		fPath := fromPath.Index(i)

		if r.Kind == "" {
			errs = append(errs, field.Required(fPath.Child("kind"), "env reference kind is required"))
		} else if r.Kind != "ConfigMap" && r.Kind != "Secret" {
			errs = append(errs, field.Invalid(fPath.Child("kind"), r, "env reference kind must be either ConfigMap or Secret"))
		}

		if r.Name == "" {
			errs = append(errs, field.Required(fPath.Child("name"), "missing env name"))
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

		if f.Ref == nil {
			errs = append(errs, field.Required(fPath.Child("ref"), "file reference is required"))
		} else {
			if f.Ref.Kind == "" {
				errs = append(errs, field.Required(fPath.Child("ref").Child("kind"), "file reference kind is required"))
			} else if f.Ref.Kind != "ConfigMap" && f.Ref.Kind != "Secret" {
				errs = append(errs, field.Invalid(fPath.Child("ref").Child("kind"), f, "file reference kind must be either ConfigMap or Secret"))
			}

			if f.Ref.Name == "" {
				errs = append(errs, field.Required(fPath.Child("ref").Child("name"), ""))
			}
			if f.Ref.Key == "" {
				errs = append(errs, field.Required(fPath.Child("ref").Child("key"), ""))
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

	if h.Instances.Max != nil {
		if *h.Instances.Max < h.Instances.Min {
			errs = append(errs, field.Invalid(fPath.Child("instances").Child("max"), *h.Instances.Max, "max cannot be smaller than min"))
		}
	}

	if h.CPUTarget != nil {
		if u := h.CPUTarget.Utilization; u != nil {
			if *u == 0 {
				errs = append(errs, field.Invalid(fPath.Child("cpuTarget").Child("utilization"), *h, "cannot be zero"))
			}

			if *u > 100 {
				errs = append(errs, field.Invalid(fPath.Child("cpuTarget").Child("utilization"), *h.CPUTarget.Utilization, "cannot be larger than 100"))
			}
		}
	}

	return errs
}
