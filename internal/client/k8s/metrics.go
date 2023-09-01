package k8s

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func (c *Client) ListCapsuleMetrics(ctx context.Context) (iterator.Iterator[*capsule.InstanceMetrics], error) {
	pid, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}
	ns := pid.String()

	p := iterator.NewProducer[*capsule.InstanceMetrics]()

	go func() {
		defer p.Done()

		lopts := metav1.ListOptions{} // TODO: only get capsule pods

		for {
			ml, err := c.mcs.MetricsV1beta1().PodMetricses(ns).List(ctx, lopts)
			if err != nil {
				p.Error(fmt.Errorf("could not list capsule metrics: %w", err))
				return
			}

			for _, ms := range ml.Items {
				cid, err := uuid.Parse(capsuleIDFromLabels(ms.GetLabels()))
				if err != nil {
					p.Error(err)
					return
				}

				cm := &capsule.InstanceMetrics{
					CapsuleId:      cid.String(),
					InstanceId:     ms.GetName(),
					MainContainer:  getMainContainerMetrics(ms.Containers, ms.Timestamp),
					ProxyContainer: getProxyContainerMetrics(ms.Containers, ms.Timestamp),
				}
				if err := p.Value(cm); err != nil {
					p.Error(err)
					return
				}
			}

			if ml.Continue == "" {
				break
			}
			lopts.Continue = ml.Continue
		}
	}()

	return p, nil
}

func getMainContainerMetrics(metrics []metricsv1beta1.ContainerMetrics, timestamp metav1.Time) *capsule.ContainerMetrics {
	for _, m := range metrics {
		if m.Name != proxyContainerName {
			return &capsule.ContainerMetrics{
				Timestamp:    timestamppb.New(timestamp.Time),
				MemoryBytes:  uint64(m.Usage.Memory().Value()),
				CpuMs:        uint64(m.Usage.Cpu().MilliValue()),
				StorageBytes: uint64(m.Usage.StorageEphemeral().MilliValue()),
			}
		}
	}
	return nil
}

func getProxyContainerMetrics(metrics []metricsv1beta1.ContainerMetrics, timestamp metav1.Time) *capsule.ContainerMetrics {
	for _, m := range metrics {
		if m.Name == proxyContainerName {
			return &capsule.ContainerMetrics{
				Timestamp:    timestamppb.New(timestamp.Time),
				MemoryBytes:  uint64(m.Usage.Memory().Value()),
				CpuMs:        uint64(m.Usage.Cpu().MilliValue()),
				StorageBytes: uint64(m.Usage.StorageEphemeral().MilliValue()),
			}
		}
	}
	return nil
}
