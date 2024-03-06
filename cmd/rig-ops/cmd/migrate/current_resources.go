package migrate

import (
	"fmt"
	"strings"

	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rivo/tview"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CurrentResources struct {
	Deployment     *appsv1.Deployment
	ServiceAccount *corev1.ServiceAccount
	Capsule        *v1alpha2.Capsule
	HPA            *autoscalingv2.HorizontalPodAutoscaler
	ConfigMaps     map[string]*corev1.ConfigMap
	Secrets        map[string]*corev1.Secret
	Services       map[string]*corev1.Service
	Ingresses      map[string]*netv1.Ingress
	CronJobs       map[string]*batchv1.CronJob
}

func NewCurrentResources() *CurrentResources {
	return &CurrentResources{
		ConfigMaps: map[string]*corev1.ConfigMap{},
		Secrets:    map[string]*corev1.Secret{},
		Services:   map[string]*corev1.Service{},
		Ingresses:  map[string]*netv1.Ingress{},
		CronJobs:   map[string]*batchv1.CronJob{},
	}
}

func (c *CurrentResources) getCurrentObject(kind, name string) client.Object {
	switch kind {
	case "Deployment":
		if d := c.Deployment; d != nil {
			c.Deployment = nil
			return d
		}
	case "HorizontalPodAutoscaler":
		if hpa := c.HPA; hpa != nil {
			c.HPA = nil
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

		if cm, ok := c.ConfigMaps[name]; ok {
			delete(c.ConfigMaps, name)
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

		if s, ok := c.Secrets[name]; ok {
			delete(c.Secrets, name)
			return s
		}
	case "Service":
		if s, ok := c.Services[name]; ok {
			delete(c.Services, name)
			return s
		}
	case "Ingress":
		if i, ok := c.Ingresses[name]; ok {
			delete(c.Ingresses, name)
			return i
		}
	case "CronJob":
		if cj, ok := c.CronJobs[name]; ok {
			delete(c.CronJobs, name)
			return cj
		}
	case "ServiceAccount":
		if sa := c.ServiceAccount; sa != nil {
			c.ServiceAccount = nil
			return sa
		}
	case "Capsule":
		if ca := c.Capsule; ca != nil {
			c.Capsule = nil
			ca.Status = nil
			return ca
		}
	}

	return nil
}

func (c *CurrentResources) CreateOverview() *tview.TreeView {
	deploymentRoot := fmt.Sprintf("Deployment/%s", c.Deployment.GetName())
	root := tview.NewTreeNode(deploymentRoot).SetSelectable(false)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	return tree
}

func (c *CurrentResources) AddResource(kind, name string, object client.Object) error {
	switch kind {
	case "Deployment":
		if c.Deployment != nil {
			return errors.New("deployment already set in CurrentResources")
		}
		d, err := convertResource[*appsv1.Deployment](object, kind)
		if err != nil {
			return err
		}
		c.Deployment = d
	case "HorizontalPodAutoscaler":
		if c.HPA != nil {
			return errors.New("horizontal pod autoscaler already set in CurrentResources")
		}
		hpa, err := convertResource[*autoscalingv2.HorizontalPodAutoscaler](object, kind)
		if err != nil {
			return err
		}
		c.HPA = hpa
	case "ConfigMap":
		if _, ok := c.ConfigMaps[name]; ok {
			return fmt.Errorf("ConfigMap '%s' already set in CurrentResources", name)
		}
		cm, err := convertResource[*corev1.ConfigMap](object, kind)
		if err != nil {
			return err
		}
		c.ConfigMaps[name] = cm
	case "Secret":
		if _, ok := c.Secrets[name]; ok {
			return fmt.Errorf("Secret '%s' already set in CurrentResources", name)
		}
		cm, err := convertResource[*corev1.Secret](object, kind)
		if err != nil {
			return err
		}
		c.Secrets[name] = cm
	case "Service":
		if _, ok := c.Services[name]; ok {
			return fmt.Errorf("Service '%s' already set in CurrentResources", name)
		}
		cm, err := convertResource[*corev1.Service](object, kind)
		if err != nil {
			return err
		}
		c.Services[name] = cm
	case "Ingress":
		if _, ok := c.Ingresses[name]; ok {
			return fmt.Errorf("Ingress '%s' already set in CurrentResources", name)
		}
		cm, err := convertResource[*netv1.Ingress](object, kind)
		if err != nil {
			return err
		}
		c.Ingresses[name] = cm
	case "CronJob":
		if _, ok := c.CronJobs[name]; ok {
			return fmt.Errorf("CronJobs '%s' already set in CurrentResources", name)
		}
		cm, err := convertResource[*batchv1.CronJob](object, kind)
		if err != nil {
			return err
		}
		c.CronJobs[name] = cm
	case "ServiceAccount":
		if c.ServiceAccount != nil {
			return errors.New("ServiceAccount already set in CurrentResources")
		}
		cm, err := convertResource[*corev1.ServiceAccount](object, kind)
		if err != nil {
			return err
		}
		c.ServiceAccount = cm
	case "Capsule":
		if c.Capsule != nil {
			return errors.New("Capsule already set in CurrentResources")
		}
		cm, err := convertResource[*v1alpha2.Capsule](object, kind)
		if err != nil {
			return err
		}
		c.Capsule = cm
	default:
		return fmt.Errorf("unexpected kind '%s' to CurrentResources", kind)
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
