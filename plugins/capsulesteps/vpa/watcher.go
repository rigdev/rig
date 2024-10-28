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
	_ plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	vpa := obj.(*vpav1.VerticalPodAutoscaler)

	rec := &apipipeline.VerticalPodAutoscalerStatus{}
	if r := vpa.Status.Recommendation; r != nil {
		for _, c := range r.ContainerRecommendations {
			// TODO Assume the main container name is the same as the VPA name (which is the capsule name).
			// We should have a capsule reference here
			if c.ContainerName == vpa.Name {
				rec.CpuMillis = &apipipeline.Recommendation{}
				if r, ok := c.Target[corev1.ResourceCPU]; ok {
					rec.CpuMillis.Target = uint64(r.MilliValue())
				}
				if r, ok := c.LowerBound[corev1.ResourceCPU]; ok {
					rec.CpuMillis.LowerBound = uint64(r.MilliValue())
				}
				if r, ok := c.UpperBound[corev1.ResourceCPU]; ok {
					rec.CpuMillis.UpperBound = uint64(r.MilliValue())
				}

				rec.MemoryBytes = &apipipeline.Recommendation{}
				if r, ok := c.Target[corev1.ResourceMemory]; ok {
					rec.MemoryBytes.Target = uint64(r.AsApproximateFloat64())
				}
				if r, ok := c.LowerBound[corev1.ResourceMemory]; ok {
					rec.MemoryBytes.LowerBound = uint64(r.AsApproximateFloat64())
				}
				if r, ok := c.UpperBound[corev1.ResourceMemory]; ok {
					rec.MemoryBytes.UpperBound = uint64(r.AsApproximateFloat64())
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
