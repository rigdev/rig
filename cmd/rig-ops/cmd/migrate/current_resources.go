package migrate

import (
	"fmt"
	"strings"

	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rivo/tview"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resources struct {
	Deployment     *appsv1.Deployment
	ServiceAccount *corev1.ServiceAccount
	Capsule        *v1alpha2.Capsule
	HPA            *autoscalingv2.HorizontalPodAutoscaler
	Service        *corev1.Service
	ConfigMaps     map[string]*corev1.ConfigMap
	Secrets        map[string]*corev1.Secret
	Ingresses      map[string]*netv1.Ingress
	CronJobs       map[string]*batchv1.CronJob
}

func NewResources() *Resources {
	return &Resources{
		ConfigMaps: map[string]*corev1.ConfigMap{},
		Secrets:    map[string]*corev1.Secret{},
		Ingresses:  map[string]*netv1.Ingress{},
		CronJobs:   map[string]*batchv1.CronJob{},
	}
}

func (r *Resources) getObject(kind, name string) client.Object {
	switch kind {
	case "Deployment":
		if d := r.Deployment; d != nil {
			r.Deployment = nil
			return d
		}
	case "HorizontalPodAutoscaler":
		if hpa := r.HPA; hpa != nil {
			r.HPA = nil
			return hpa
		}
	case "ConfigMap":
		parts := strings.Split(name, "--")
		if len(parts) < 2 {
			name = fmt.Sprintf("env-source--%s", name)
		} else {
			name = strings.Replace("/"+parts[1], "-", "/", -1)
			i := strings.LastIndex(name, "/")
			name = name[:i] + "." + name[i+1:]
		}

		if cm, ok := r.ConfigMaps[name]; ok {
			delete(r.ConfigMaps, name)
			return cm
		}
	case "Secret":
		parts := strings.Split(name, "--")
		if len(parts) < 2 {
			name = fmt.Sprintf("env-source--%s", name)
		} else {
			name = strings.Replace("/"+parts[1], "-", "/", -1)
			i := strings.LastIndex(name, "/")
			name = name[:i] + "." + name[i+1:]
		}

		if s, ok := r.Secrets[name]; ok {
			delete(r.Secrets, name)
			return s
		}
	case "Service":
		if s := r.Service; s != nil {
			r.Service = nil
			return s
		}
	case "Ingress":
		if i, ok := r.Ingresses[name]; ok {
			delete(r.Ingresses, name)
			return i
		}
	case "CronJob":
		if cj, ok := r.CronJobs[name]; ok {
			delete(r.CronJobs, name)
			return cj
		}
	case "ServiceAccount":
		if sa := r.ServiceAccount; sa != nil {
			r.ServiceAccount = nil
			return sa
		}
	case "Capsule":
		if ca := r.Capsule; ca != nil {
			r.Capsule = nil
			ca.Status = nil
			return ca
		}
	}

	return nil
}

func (r *Resources) CreateOverview() *tview.TreeView {
	deploymentRoot := fmt.Sprintf("Deployment/%s", r.Deployment.GetName())
	root := tview.NewTreeNode(deploymentRoot).SetSelectable(false)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	add := func(parent *tview.TreeNode, kind string, name string) *tview.TreeNode {
		node := tview.NewTreeNode(fmt.Sprintf("%s/%s", kind, name)).
			SetSelectable(false)

		parent.AddChild(node)
		return node
	}

	if r.ServiceAccount != nil {
		add(root, "ServiceAccount", r.ServiceAccount.GetName())
	}

	if r.HPA != nil {
		add(root, "HorizontalPodAutoscaler", r.HPA.GetName())
	}

	for _, c := range r.ConfigMaps {
		add(root, "ConfigMap", c.GetName())
	}

	for _, s := range r.Secrets {
		add(root, "Secret", s.GetName())
	}

	if r.Service != nil {
		serviceNode := add(root, "Service", r.Service.GetName())

		for name, i := range r.Ingresses {
			for _, p := range i.Spec.Rules[0].HTTP.Paths {
				if p.Backend.Service.Name == name {
					add(serviceNode, "Ingress", i.GetName())
				}
			}
		}

	}

	tree.Box.SetTitle("Current Resources (ESC to exit)").
		SetBorder(true)

	return tree
}

func (r *Resources) Compare(other *Resources, scheme *runtime.Scheme) (*ReportSet, error) {
	reportSet := NewReportSet(scheme)

	if r.Deployment != nil {
		originalDeployment := other.getObject("Deployment", r.Deployment.Name)
		if err := reportSet.AddReport(originalDeployment, r.Deployment); err != nil {
			return nil, err
		}
	}

	if r.HPA != nil {
		originalHPA := other.getObject("HorizontalPodAutoscaler", r.HPA.Name)
		if err := reportSet.AddReport(originalHPA, r.HPA); err != nil {
			return nil, err
		}
	}

	if r.ServiceAccount != nil {
		originalServiceAccount := other.getObject("ServiceAccount", r.ServiceAccount.Name)
		if err := reportSet.AddReport(originalServiceAccount, r.ServiceAccount); err != nil {
			return nil, err
		}
	}

	if r.Service != nil {
		originalService := other.getObject("Service", r.Service.Name)
		if err := reportSet.AddReport(originalService, r.Service); err != nil {
			return nil, err
		}
	}

	if r.Capsule != nil {
		originalCapsule := other.getObject("Capsule", r.Capsule.Name)
		if err := reportSet.AddReport(originalCapsule, r.Capsule); err != nil {
			return nil, err
		}
	}

	for _, configMap := range r.ConfigMaps {
		originalConfigMap := other.getObject("ConfigMap", configMap.Name)
		if err := reportSet.AddReport(originalConfigMap, configMap); err != nil {
			return nil, err
		}
	}

	for _, secret := range r.Secrets {
		originalSecret := other.getObject("Secret", secret.Name)
		if err := reportSet.AddReport(originalSecret, secret); err != nil {
			return nil, err
		}
	}

	for _, ingress := range r.Ingresses {
		originalIngress := other.getObject("Ingress", ingress.Name)
		if err := reportSet.AddReport(originalIngress, ingress); err != nil {
			return nil, err
		}
	}

	for _, cronJob := range r.CronJobs {
		originalCronJob := other.getObject("CronJob", cronJob.Name)
		if err := reportSet.AddReport(originalCronJob, cronJob); err != nil {
			return nil, err
		}
	}

	return reportSet, nil
}

func (r *Resources) AddObject(kind, name string, object client.Object) error {
	switch kind {
	case "Deployment":
		if r.Deployment != nil {
			return errors.AlreadyExistsErrorf("deployment already set in current resources")
		}
		d, err := convertResource[*appsv1.Deployment](object, kind)
		if err != nil {
			return err
		}
		d.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind(kind))
		r.Deployment = d
	case "HorizontalPodAutoscaler":
		if r.HPA != nil {
			return errors.AlreadyExistsErrorf("horizontal pod autoscaler already set in current resources")
		}
		hpa, err := convertResource[*autoscalingv2.HorizontalPodAutoscaler](object, kind)
		if err != nil {
			return err
		}
		hpa.SetGroupVersionKind(autoscalingv2.SchemeGroupVersion.WithKind(kind))
		r.HPA = hpa

	case "Service":
		if r.Service != nil {
			return errors.AlreadyExistsErrorf("service '%s' already set in current resources", name)
		}
		s, err := convertResource[*corev1.Service](object, kind)
		if err != nil {
			return err
		}
		s.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind(kind))
		r.Service = s
	case "ConfigMap":
		if _, ok := r.ConfigMaps[name]; ok {
			return errors.AlreadyExistsErrorf("config map '%s' already set in current resources", name)
		}
		cm, err := convertResource[*corev1.ConfigMap](object, kind)
		if err != nil {
			return err
		}
		cm.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind(kind))
		r.ConfigMaps[name] = cm
	case "Secret":
		if _, ok := r.Secrets[name]; ok {
			return errors.AlreadyExistsErrorf("secret '%s' already set in current resources", name)
		}
		s, err := convertResource[*corev1.Secret](object, kind)
		if err != nil {
			return err
		}
		s.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind(kind))
		r.Secrets[name] = s
	case "Ingress":
		if _, ok := r.Ingresses[name]; ok {
			return errors.AlreadyExistsErrorf("ingress '%s' already set in current resources", name)
		}
		i, err := convertResource[*netv1.Ingress](object, kind)
		if err != nil {
			return err
		}
		i.SetGroupVersionKind(netv1.SchemeGroupVersion.WithKind(kind))
		r.Ingresses[name] = i
	case "CronJob":
		if _, ok := r.CronJobs[name]; ok {
			return errors.AlreadyExistsErrorf("cron jobs '%s' already set in current resources", name)
		}
		cj, err := convertResource[*batchv1.CronJob](object, kind)
		if err != nil {
			return err
		}
		cj.SetGroupVersionKind(batchv1.SchemeGroupVersion.WithKind(kind))
		r.CronJobs[name] = cj
	case "ServiceAccount":
		if r.ServiceAccount != nil {
			return errors.AlreadyExistsErrorf("service account already set in current resources")
		}
		s, err := convertResource[*corev1.ServiceAccount](object, kind)
		if err != nil {
			return err
		}
		s.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind(kind))
		r.ServiceAccount = s
	case "Capsule":
		if r.Capsule != nil {
			return errors.AlreadyExistsErrorf("capsule already set in current resources")
		}
		ca, err := convertResource[*v1alpha2.Capsule](object, kind)
		if err != nil {
			return err
		}
		ca.SetGroupVersionKind(v1alpha2.GroupVersion.WithKind(kind))
		r.Capsule = ca
	default:
		return errors.InvalidArgumentErrorf("unexpected kind '%s' to current resources", kind)
	}

	return nil
}

func convertResource[T any](object client.Object, kind string) (T, error) {
	var empty T
	d, ok := object.(T)
	if !ok {
		return empty, fmt.Errorf("kind was %s, but type was %T", kind, object)
	}
	return d, nil
}

func (r *Resources) ToYAML(scheme *runtime.Scheme) (map[string]map[string]string, error) {
	res := map[string]map[string]string{}
	var err error

	if r.Deployment != nil {
		res["Deployment"] = map[string]string{}
		res["Deployment"][r.Deployment.Name], err = toYamlString(r.Deployment, scheme)
		if err != nil {
			return nil, err
		}
	}

	if r.HPA != nil {
		res["HorizontalPodAutoscaler"] = map[string]string{}
		res["HorizontalPodAutoscaler"][r.HPA.Name], err = toYamlString(r.HPA, scheme)
		if err != nil {
			return nil, err
		}
	}

	if r.ServiceAccount != nil {
		res["ServiceAccount"] = map[string]string{}
		res["ServiceAccount"][r.ServiceAccount.Name], err = toYamlString(r.ServiceAccount, scheme)
		if err != nil {
			return nil, err
		}
	}

	if len(r.ConfigMaps) > 0 {
		res["ConfigMap"] = map[string]string{}
	}
	for _, configMap := range r.ConfigMaps {
		res["ConfigMap"][configMap.Name], err = toYamlString(configMap, scheme)
		if err != nil {
			return nil, err
		}
	}

	if len(r.Secrets) > 0 {
		res["Secret"] = map[string]string{}
	}
	for _, secret := range r.Secrets {
		res["Secret"][secret.Name], err = toYamlString(secret, scheme)
		if err != nil {
			return nil, err
		}
	}

	if r.Service != nil {
		res["Service"] = map[string]string{}
		res["Service"][r.Service.Name], err = toYamlString(r.Service, scheme)
		if err != nil {
			return nil, err
		}

	}

	if len(r.Ingresses) > 0 {
		res["Ingress"] = map[string]string{}
	}
	for _, ingress := range r.Ingresses {
		res["Ingress"][ingress.Name], err = toYamlString(ingress, scheme)
		if err != nil {
			return nil, err
		}
	}

	if len(r.CronJobs) > 0 {
		res["CronJob"] = map[string]string{}
	}

	for _, cronJob := range r.CronJobs {
		res["CronJob"][cronJob.Name], err = toYamlString(cronJob, scheme)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func toYamlString(object client.Object, scheme *runtime.Scheme) (string, error) {
	object.SetManagedFields(nil)
	bs, err := obj.Encode(object, scheme)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}
