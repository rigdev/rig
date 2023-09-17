package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/pkg/utils"
)

func getCurrentContainerSettings(ctx context.Context, capsuleID CapsuleID, client rig.Client) (*capsule.ContainerSettings, error) {
	resp, err := client.Capsule().ListRollouts(ctx, connect.NewRequest(&capsule.ListRolloutsRequest{
		CapsuleId: capsuleID,
		Pagination: &model.Pagination{
			Offset:     0,
			Limit:      1,
			Descending: true,
		},
	}))
	if err != nil {
		return nil, err
	}

	if resp.Msg.Total == 0 {
		return nil, nil
	}

	r, err := client.Capsule().GetRollout(ctx, connect.NewRequest(&capsule.GetRolloutRequest{
		CapsuleId: capsuleID,
		RolloutId: resp.Msg.Rollouts[0].RolloutId,
	}))
	if err != nil {
		return nil, err
	}

	container := r.Msg.GetRollout().GetConfig().GetContainerSettings()
	if container == nil {
		container = &capsule.ContainerSettings{}
	}
	if container.Resources == nil {
		container.Resources = &capsule.Resources{}
	}

	utils.FeedDefaultResources(container.Resources)

	return container, nil
}
