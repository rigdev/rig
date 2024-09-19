package argorollout

import (
	"context"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/plugins/capsulesteps/deployment"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func onRolloutUpdated(obj client.Object, _ []*corev1.Event, watcher plugin.ObjectWatcher) *apipipeline.ObjectStatusInfo {
	// TODO Observe Argo Rollout specific conditions as well to monitor the rollout
	rollout := obj.(*v1alpha1.Rollout)
	return deployment.OnPodTemplatedUpdated(rollout.Spec.Template, watcher)
}

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	return watcher.WatchPrimary(ctx, &v1alpha1.Rollout{}, onRolloutUpdated)
}
