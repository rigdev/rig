package migrate

import (
	"fmt"
	"strings"

	"github.com/rigdev/rig/pkg/api/v1alpha2"
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

func (c *CurrentResources) createOverview() *tview.TreeView {
	deploymentRoot := fmt.Sprintf("Deployment/%s", c.Deployment.GetName())
	root := tview.NewTreeNode(deploymentRoot).SetSelectable(false)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	return tree
}
