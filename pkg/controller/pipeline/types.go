package pipeline

import (
	"fmt"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime/schema"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	AppsDeploymentGVK                      = appsv1.SchemeGroupVersion.WithKind("Deployment")
	CoreServiceGVK                         = corev1.SchemeGroupVersion.WithKind("Service")
	CMCertificateGVK                       = cmv1.SchemeGroupVersion.WithKind(cmv1.CertificateKind)
	NetIngressGVK                          = netv1.SchemeGroupVersion.WithKind("Ingress")
	AutoscalingvHorizontalPodAutoscalerGVK = autoscalingv2.SchemeGroupVersion.WithKind("HorizontalPodAutoscaler")
	BatchCronJobGVK                        = batchv1.SchemeGroupVersion.WithKind("CronJob")
	MonitoringServiceMonitorGVK            = monitorv1.SchemeGroupVersion.WithKind(monitorv1.ServiceMonitorsKind)
	VPAVerticalPodAutoscalerGVK            = vpav1.SchemeGroupVersion.WithKind("VerticalPodAutoscaler")
	CoreServiceAccount                     = corev1.SchemeGroupVersion.WithKind("ServiceAccount")

	_allGVKs = []schema.GroupVersionKind{
		AppsDeploymentGVK,
		CoreServiceGVK,
		CMCertificateGVK,
		NetIngressGVK,
		AutoscalingvHorizontalPodAutoscalerGVK,
		BatchCronJobGVK,
		MonitoringServiceMonitorGVK,
		VPAVerticalPodAutoscalerGVK,
		CoreServiceAccount,
	}

	_gvkByAPIGroupKind = map[string]map[string]schema.GroupVersionKind{}
	// If the capsule was created before the pipeline refactor then there isn't a
	// group associated to the resources in OwnedResources.
	_gvkByKind = map[string]schema.GroupVersionKind{}
)

func init() {
	for _, gvk := range _allGVKs {
		gs, ok := _gvkByAPIGroupKind[gvk.Group]
		if !ok {
			gs = map[string]schema.GroupVersionKind{}
			_gvkByAPIGroupKind[gvk.Group] = gs
		}

		gs[gvk.Kind] = gvk

		_gvkByKind[gvk.Kind] = gvk
	}
}

func LookupGVK(gk schema.GroupKind) (schema.GroupVersionKind, error) {
	if gk.Group == "" {
		gvk, ok := _gvkByKind[gk.Kind]
		if !ok {
			return schema.GroupVersionKind{}, fmt.Errorf("unknown kind '%v' with empty apiGroup", gk.Kind)
		}
		return gvk, nil
	}

	gs, ok := _gvkByAPIGroupKind[gk.Group]
	if !ok {
		return schema.GroupVersionKind{}, fmt.Errorf("unknown apiGroup '%v'", gk.Group)
	}

	gvk, ok := gs[gk.Kind]
	if !ok {
		return schema.GroupVersionKind{}, fmt.Errorf("unknown kind '%v' in apiGroup '%v'", gk.Kind, gk.Group)
	}

	return gvk, nil
}

type ObjectsEqual func(o1, o2 client.Object) bool

var _objectsEquals = map[schema.GroupVersionKind]ObjectsEqual{
	MonitoringServiceMonitorGVK: func(o1, o2 client.Object) bool {
		return equality.Semantic.DeepEqual(o1.(*monitorv1.ServiceMonitor).Spec, o2.(*monitorv1.ServiceMonitor).Spec)
	},
	AppsDeploymentGVK: func(o1, o2 client.Object) bool {
		return equality.Semantic.DeepEqual(o1.(*appsv1.Deployment).Spec, o2.(*appsv1.Deployment).Spec)
	},
}

func ObjectsEquals(o1, o2 client.Object) bool {
	objectsEqual, ok := _objectsEquals[o1.GetObjectKind().GroupVersionKind()]
	if !ok {
		objectsEqual = func(o1, o2 client.Object) bool {
			return equality.Semantic.DeepEqual(o1, o2)
		}
	}

	return objectsEqual(o1, o2)
}
