package root

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c *Cmd) abort(ctx context.Context, cmd *cobra.Command, _ []string) error {
	cc, err := c.Rig.Capsule().Get(ctx, &connect.Request[capsule.GetRequest]{
		Msg: &capsule.GetRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			ProjectId: c.Cfg.GetProject(),
		},
	})
	if err != nil {
		return err
	}

	if _, err := c.Rig.Capsule().AbortRollout(ctx, &connect.Request[capsule.AbortRolloutRequest]{
		Msg: &capsule.AbortRolloutRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			RolloutId: cc.Msg.GetCapsule().GetCurrentRollout(),
			ProjectId: c.Cfg.GetProject(),
		},
	}); err != nil {
		return err
	}

	cmd.Println("Current rollout aborted")

	return nil
}
