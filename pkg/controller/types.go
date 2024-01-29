package controller

import (
	"fmt"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

var (
	_appsDeploymentGVK                      = appsv1.SchemeGroupVersion.WithKind("Deployment")
	_coreServiceGVK                         = corev1.SchemeGroupVersion.WithKind("Service")
	_cmCertificateGVK                       = cmv1.SchemeGroupVersion.WithKind(cmv1.CertificateKind)
	_netIngressGVK                          = netv1.SchemeGroupVersion.WithKind("Ingress")
	_autoscalingvHorizontalPodAutoscalerGVK = autoscalingv2.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler")
	_batchCronJobGVK                        = batchv1.SchemeGroupVersion.WithKind("CronJob")
	_monitoringServiceMonitorGVK            = monitorv1.SchemeGroupVersion.WithKind(monitorv1.ServiceMonitorsKind)
	_vpaVerticalPodAutoscalerGVK            = vpav1.SchemeGroupVersion.WithKind("VerticalPodAutoscaler")
	_coreServiceAccount                     = corev1.SchemeGroupVersion.WithKind("ServiceAccount")

	_allGVKs = []schema.GroupVersionKind{
		_appsDeploymentGVK,
		_coreServiceGVK,
		_cmCertificateGVK,
		_netIngressGVK,
		_autoscalingvHorizontalPodAutoscalerGVK,
		_batchCronJobGVK,
		_monitoringServiceMonitorGVK,
		_vpaVerticalPodAutoscalerGVK,
		_coreServiceAccount,
	}

	_gvkByApiGroupKind = map[string]map[string]schema.GroupVersionKind{}
)

func init() {
	for _, gvk := range _allGVKs {
		gs, ok := _gvkByApiGroupKind[gvk.Group]
		if !ok {
			gs = map[string]schema.GroupVersionKind{}
			_gvkByApiGroupKind[gvk.Group] = gs
		}

		gs[gvk.Kind] = gvk
	}
}

func lookupGVK(gk schema.GroupKind) (schema.GroupVersionKind, error) {
	gs, ok := _gvkByApiGroupKind[gk.Group]
	if !ok {
		return schema.GroupVersionKind{}, fmt.Errorf("unknown apiGroup '%v'", gk.Group)
	}

	gvk, ok := gs[gk.Kind]
	if !ok {
		return schema.GroupVersionKind{}, fmt.Errorf("unknown kind '%v' in apiGroup '%v'", gk.Kind, gk.Group)
	}

	return gvk, nil
}
