package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/spf13/cobra"
)

func CapsuleAbort(ctx context.Context, cmd *cobra.Command, capsuleID CapsuleID, nc rig.Client) error {
	c, err := nc.Capsule().Get(ctx, &connect.Request[capsule.GetRequest]{
		Msg: &capsule.GetRequest{
			CapsuleId: capsuleID.String(),
		},
	})
	if err != nil {
		return err
	}

	if _, err := nc.Capsule().AbortRollout(ctx, &connect.Request[capsule.AbortRolloutRequest]{
		Msg: &capsule.AbortRolloutRequest{
			CapsuleId: capsuleID.String(),
			RolloutId: c.Msg.GetCapsule().GetCurrentRollout(),
		},
	}); err != nil {
		return err
	}

	cmd.Println("Rollout aborted")

	return nil
}
