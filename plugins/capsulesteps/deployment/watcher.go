package deployment

import (
	"context"

	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func onPodUpdated(
	obj client.Object,
	events []*corev1.Event,
	_ plugin.ObjectWatcher,
) *apipipeline.ObjectStatus {
	pod := obj.(*corev1.Pod)

	status := &apipipeline.ObjectStatus{
		Type:       apipipeline.ObjectType_OBJECT_TYPE_POD,
		Properties: map[string]string{},
	}

	for _, c := range pod.Status.Conditions {
		cond := &apipipeline.ObjectCondition{
			Name:    string(c.Type),
			State:   apipipeline.ObjectState_OBJECT_STATE_PENDING,
			Message: c.Message,
		}
		status.Conditions = append(status.Conditions, cond)
	}

	for _, e := range events {
		cond := &apipipeline.ObjectCondition{
			Name:    string(e.Name),
			State:   apipipeline.ObjectState_OBJECT_STATE_PENDING,
			Message: e.Message,
		}
		status.Conditions = append(status.Conditions, cond)
	}

	return status
}

func onDeploymentUpdated(
	obj client.Object,
	_ []*corev1.Event,
	objectWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatus {
	dep := obj.(*appsv1.Deployment)

	objectWatcher.WatchSecondaryByLabels(labels.Set(dep.Spec.Template.GetLabels()), &corev1.Pod{}, onPodUpdated)

	status := &apipipeline.ObjectStatus{
		Type:       apipipeline.ObjectType_OBJECT_TYPE_PRIMARY,
		Properties: map[string]string{},
	}

	return status
}

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	return watcher.WatchPrimary(ctx, &appsv1.Deployment{}, onDeploymentUpdated)
}
