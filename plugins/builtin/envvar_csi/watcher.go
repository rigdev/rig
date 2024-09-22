package envvarcsi

import (
	"context"
	"fmt"

	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/plugins/capsulesteps/deployment"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	return watcher.WatchPrimary(ctx, &appsv1.Deployment{}, onDeploymentUpdated)
}

func onDeploymentUpdated(
	obj client.Object,
	_ []*corev1.Event,
	objectWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	dep := obj.(*appsv1.Deployment)
	objectWatcher.WatchSecondaryByLabels(deployment.PodLabelSelector(dep.Spec.Template), &corev1.Pod{}, onPodUpdated)
	return &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
	}
}

func onPodUpdated(
	obj client.Object,
	events []*corev1.Event,
	watcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
	}
	pod := obj.(*corev1.Pod)
	capsuleName := pod.Labels["rig.dev/capsule"]
	secretName := k8sSecretName(capsuleName)
	for _, c := range pod.Status.ContainerStatuses {
		cond := &apipipeline.ObjectCondition{
			Name:    "Preparing",
			State:   apipipeline.ObjectState_OBJECT_STATE_HEALTHY,
			Message: fmt.Sprintf("Secret Store CSI has created its secret"),
		}
		if w := c.State.Waiting; w != nil {
			if w.Reason == "CreateContainerConfigError" && w.Message == fmt.Sprintf("secret \"%s\" not found", secretName) {
				cond.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
				cond.Message = fmt.Sprintf("Some environment variables for container '%s' should be stored by Secret Store CSI Driver into a secret '%s', but the secret does not yet exist.", c.Name, secretName)
			}
		}
		status.SubObjects = append(status.SubObjects, &apipipeline.SubObjectStatus{
			Name:       capsuleName,
			Conditions: []*apipipeline.ObjectCondition{cond},
		})
	}

	return status
}
