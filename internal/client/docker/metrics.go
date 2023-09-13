package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) ListCapsuleMetrics(ctx context.Context) (iterator.Iterator[*capsule.InstanceMetrics], error) {
	pid, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	p := iterator.NewProducer[*capsule.InstanceMetrics]()

	go func() {
		defer p.Done()

		cs, err := c.dc.ContainerList(ctx, types.ContainerListOptions{
			All: true,
		})
		if err != nil {
			p.Error(fmt.Errorf("could not list containers: %w", err))
			return
		}

		for _, cc := range cs {
			if pidLabel, ok := cc.Labels[_rigProjectIDLabel]; !ok || pidLabel != pid.String() {
				// ignore non project containers
				continue
			}

			cidLabel, ok := cc.Labels[_rigCapsuleIDLabel]
			if !ok {
				// ignore containers without a capsule id label
				continue
			}

			ccs, err := c.dc.ContainerStatsOneShot(ctx, cc.ID)
			if err != nil {
				p.Error(fmt.Errorf("could not get container stats: %w", err))
				return
			}
			defer ccs.Body.Close()

			var s types.StatsJSON
			if err := json.NewDecoder(ccs.Body).Decode(&s); err != nil {
				p.Error(fmt.Errorf("could not decode container stats: %w", err))
				return
			}

			cm := &capsule.InstanceMetrics{
				CapsuleId:  cidLabel,
				InstanceId: containerName(cc),
				MainContainer: &capsule.ContainerMetrics{
					Timestamp:   timestamppb.New(s.Read),
					CpuMs:       s.CPUStats.CPUUsage.TotalUsage / 1e6,
					MemoryBytes: s.MemoryStats.Usage,
				},
			}

			if err := p.Value(cm); err != nil {
				p.Error(err)
				return
			}
		}
	}()

	return p, nil
}
