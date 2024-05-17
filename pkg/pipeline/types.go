package pipeline

import (
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	AppsDeploymentGVK           = appsv1.SchemeGroupVersion.WithKind("Deployment")
	MonitoringServiceMonitorGVK = monitorv1.SchemeGroupVersion.WithKind(monitorv1.ServiceMonitorsKind)
)

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
