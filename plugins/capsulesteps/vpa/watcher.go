package vpa

import (
	"context"

	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	corev1 "k8s.io/api/core/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	return watcher.WatchPrimary(ctx, &vpav1.VerticalPodAutoscaler{}, onVPAUpdated)
}

func onVPAUpdated(
	obj client.Object,
	_ []*corev1.Event,
	objWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	vpa := obj.(*vpav1.VerticalPodAutoscaler)

	rec := &apipipeline.VerticalPodAutoscalerStatus{}
	for _, c := range vpa.Status.Recommendation.ContainerRecommendations {
		// TODO Assume the main container name is the same as the VPA name (which is the capsule name).
		// We should have a capsule reference here
		if c.ContainerName == vpa.Name {
			if r, ok := c.Target[corev1.ResourceCPU]; ok {
				rec.CpuMillis = &apipipeline.Recommendation{
					Target: uint64(r.AsApproximateFloat64() * 1000.),
				}
			}
			if r, ok := c.Target[corev1.ResourceMemory]; ok {
				rec.MemoryBytes = &apipipeline.Recommendation{
					Target: uint64(r.AsApproximateFloat64()),
				}
			}
		}
	}

	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
		PlatformStatus: []*apipipeline.PlatformObjectStatus{
			{
				Name: vpa.Name,
				Kind: &apipipeline.PlatformObjectStatus_Vpa{
					Vpa: rec,
				},
			},
		},
	}

	return status
}
