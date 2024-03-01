package migrate

import (
	"encoding/json"
	"strings"

	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
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

func (c *CurrentResources) getCurrentObject(kind, name string) ([]byte, error) {
	var orig []byte
	var err error
	switch kind {
	case "Deployment":
		if c.Deployment != nil {
			c.Deployment.SetManagedFields(nil)
			c.Deployment.Status = appsv1.DeploymentStatus{}
			orig, err = json.Marshal(c.Deployment)
			if err != nil {
				return nil, err
			}
			c.Deployment = nil
		}
	case "HorizontalPodAutoscaler":
		if c.HPA != nil {
			c.HPA.SetManagedFields(nil)
			c.HPA.Status = autoscalingv2.HorizontalPodAutoscalerStatus{}
			orig, err = json.Marshal(c.HPA)
			if err != nil {
				return nil, err
			}
			c.HPA = nil
		}
	case "ConfigMap":
		parts := strings.Split(name, "--")
		if len(parts) < 2 {
			name = "env-source"
		} else {
			name = strings.Replace("/"+parts[1], "-", "/", -1)
			i := strings.LastIndex(name, "/")
			name = name[:i] + "." + name[i+1:]
		}

		if cm, ok := c.ConfigMaps[name]; ok {
			cm.SetManagedFields(nil)
			orig, err = json.Marshal(cm)
			if err != nil {
				return nil, err
			}

			delete(c.ConfigMaps, name)
		}
	case "Secret":
		parts := strings.Split(name, "--")
		if len(parts) < 2 {
			name = "env-source"
		} else {
			name = strings.Replace("/"+parts[1], "-", "/", -1)
			i := strings.LastIndex(name, "/")
			name = name[:i] + "." + name[i+1:]
		}

		if s, ok := c.Secrets[name]; ok {
			s.SetManagedFields(nil)
			orig, err = json.Marshal(s)
			if err != nil {
				return nil, err
			}

			delete(c.Secrets, name)
		}
	case "Service":
		if s, ok := c.Services[name]; ok {
			s.SetManagedFields(nil)
			orig, err = json.Marshal(s)
			if err != nil {
				return nil, err
			}

			delete(c.Services, name)
		}
	case "Ingress":
		if i, ok := c.Ingresses[name]; ok {
			i.SetManagedFields(nil)
			orig, err = json.Marshal(i)
			if err != nil {
				return nil, err
			}

			delete(c.Ingresses, name)
		}
	case "CronJob":
		if cj, ok := c.CronJobs[name]; ok {
			cj.SetManagedFields(nil)
			orig, err = json.Marshal(cj)
			if err != nil {
				return nil, err
			}

			delete(c.CronJobs, name)
		}
	case "ServiceAccount":
		if sa := c.ServiceAccount; sa != nil {
			sa.SetManagedFields(nil)
			orig, err = json.Marshal(sa)
			if err != nil {
				return nil, err
			}

			c.ServiceAccount = nil
		}
	case "Capsule":
		if c.Capsule != nil {
			c.Capsule.Status = nil
			c.Capsule.SetManagedFields(nil)
			orig, err = json.Marshal(c.Capsule)
			if err != nil {
				return nil, err
			}

			c.Capsule = nil
		}
	}

	return orig, nil
}

func (c *CurrentResources) ToYAML(cc client.Client) (map[string]string, error) {
	deploymentCopy := c.Deployment.DeepCopy()
	deploymentCopy.ManagedFields = nil

	deploymentYAML, err := obj.Encode(deploymentCopy, cc.Scheme())
	if err != nil {
		return nil, err
	}

	configMapList := &corev1.ConfigMapList{}
	for _, configMap := range c.ConfigMaps {
		configMapCopy := configMap.DeepCopy()
		configMapCopy.ManagedFields = nil
		configMapList.Items = append(configMapList.Items, *configMapCopy)
	}
	configMapsYAML, err := obj.Encode(configMapList, cc.Scheme())
	if err != nil {
		return nil, err
	}

	secretList := &corev1.SecretList{}
	for _, secret := range c.Secrets {
		secretCopy := secret.DeepCopy()
		secretCopy.ManagedFields = nil
		secretList.Items = append(secretList.Items, *secretCopy)
	}
	secretsYAML, err := obj.Encode(secretList, cc.Scheme())
	if err != nil {
		return nil, err
	}

	serviceList := &corev1.ServiceList{}
	for _, service := range c.Services {
		serviceCopy := service.DeepCopy()
		serviceCopy.ManagedFields = nil
		serviceList.Items = append(serviceList.Items, *serviceCopy)
	}
	servicesYAML, err := obj.Encode(serviceList, cc.Scheme())
	if err != nil {
		return nil, err
	}

	ingressList := &netv1.IngressList{}
	for _, ingress := range c.Ingresses {
		ingressCopy := ingress.DeepCopy()
		ingressCopy.ManagedFields = nil
		ingressList.Items = append(ingressList.Items, *ingressCopy)
	}
	ingressesYAML, err := obj.Encode(ingressList, cc.Scheme())
	if err != nil {
		return nil, err
	}

	cronJobList := &batchv1.CronJobList{}
	for _, cronJob := range c.CronJobs {
		cronJobCopy := cronJob.DeepCopy()
		cronJobCopy.ManagedFields = nil
		cronJobList.Items = append(cronJobList.Items, *cronJobCopy)
	}
	cronJobsYAML, err := obj.Encode(cronJobList, cc.Scheme())
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"deployment": string(deploymentYAML),
		"configMaps": string(configMapsYAML),
		"secrets":    string(secretsYAML),
		"services":   string(servicesYAML),
		"ingresses":  string(ingressesYAML),
		"cronJobs":   string(cronJobsYAML),
	}, nil
}
