package capsule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
)

var CapsuleID string

func GetCurrentContainerResources(ctx context.Context, client rig.Client) (*capsule.ContainerSettings, uint32, error) {
	resp, err := client.Capsule().Get(ctx, connect.NewRequest(&capsule.GetRequest{
		CapsuleId: CapsuleID,
	}))
	if err != nil {
		return nil, 0, err
	}

	r, err := client.Capsule().GetRollout(ctx, connect.NewRequest(&capsule.GetRolloutRequest{
		CapsuleId: CapsuleID,
		RolloutId: resp.Msg.GetCapsule().GetCurrentRollout(),
	}))
	if err != nil {
		return nil, 0, err
	}

	container := r.Msg.GetRollout().GetConfig().GetContainerSettings()
	if container == nil {
		container = &capsule.ContainerSettings{}
	}
	if container.Resources == nil {
		container.Resources = &capsule.Resources{}
	}

	utils.FeedDefaultResources(container.Resources)

	return container, r.Msg.GetRollout().GetConfig().GetReplicas(), nil
}

func GetCurrentNetwork(ctx context.Context, client rig.Client) (*capsule.Network, error) {
	resp, err := client.Capsule().Get(ctx, connect.NewRequest(&capsule.GetRequest{
		CapsuleId: CapsuleID,
	}))
	if err != nil {
		return nil, err
	}

	r, err := client.Capsule().GetRollout(ctx, connect.NewRequest(&capsule.GetRolloutRequest{
		CapsuleId: CapsuleID,
		RolloutId: resp.Msg.GetCapsule().GetCurrentRollout(),
	}))
	if err != nil {
		return nil, err
	}

	return r.Msg.GetRollout().GetConfig().GetNetwork(), nil
}

func GetCurrentRollout(ctx context.Context, client rig.Client) (*capsule.Rollout, error) {
	resp, err := client.Capsule().Get(ctx, connect.NewRequest(&capsule.GetRequest{
		CapsuleId: CapsuleID,
	}))
	if err != nil {
		return nil, err
	}

	r, err := client.Capsule().GetRollout(ctx, connect.NewRequest(&capsule.GetRolloutRequest{
		CapsuleId: CapsuleID,
		RolloutId: resp.Msg.GetCapsule().GetCurrentRollout(),
	}))
	if err != nil {
		return nil, err
	}

	return r.Msg.GetRollout(), nil
}

func formatCapsule(c *capsule.Capsule) string {
	var age string
	if c.GetCurrentRollout() == 0 {
		age = "-"
	} else {
		age = time.Since(c.GetUpdatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Rollout: %v, Updated At: %v)", c.GetCapsuleId(), c.GetCurrentRollout(), age)
}

func Truncated(str string, max int) string {
	if len(str) > max {
		return str[:strings.LastIndexAny(str[:max], " .,:;-")] + "..."
	}

	return str
}

func TruncatedFixed(str string, max int) string {
	if len(str) > max {
		return str[:max] + "..."
	}

	return str
}

func AbortAndDeploy(ctx context.Context, capsuleID string, rig rig.Client, req *connect.Request[capsule.DeployRequest]) (*connect.Response[capsule.DeployResponse], error) {
	deploy, err := common.PromptConfirm("Rollout already in progress, would you like to cancel it and redeploy?", false)
	if err != nil {
		return nil, err
	}

	if !deploy {
		return nil, errors.FailedPreconditionErrorf("rollout already in progress")
	}

	cc, err := rig.Capsule().Get(ctx, &connect.Request[capsule.GetRequest]{
		Msg: &capsule.GetRequest{
			CapsuleId: capsuleID,
		},
	})
	if err != nil {
		return nil, err
	}

	if _, err := rig.Capsule().AbortRollout(ctx, &connect.Request[capsule.AbortRolloutRequest]{
		Msg: &capsule.AbortRolloutRequest{
			CapsuleId: capsuleID,
			RolloutId: cc.Msg.GetCapsule().GetCurrentRollout(),
		},
	}); err != nil {
		return nil, err
	}

	return rig.Capsule().Deploy(ctx, req)

}
