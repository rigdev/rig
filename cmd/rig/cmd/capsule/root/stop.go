package root

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c *Cmd) stop(ctx context.Context, cmd *cobra.Command, _ []string) error {
	currentRollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Scope)
	if err != nil {
		return err
	}

	if _, err := c.Rig.Capsule().StopRollout(ctx, &connect.Request[capsule.StopRolloutRequest]{
		Msg: &capsule.StopRolloutRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			RolloutId: currentRollout.GetRolloutId(),
			ProjectId: c.Scope.GetCurrentContext().GetProject(),
		},
	}); err != nil {
		return err
	}

	cmd.Println("rollout stopped")

	return nil
}
