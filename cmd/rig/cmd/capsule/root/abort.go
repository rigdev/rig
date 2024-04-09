package root

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) abort(ctx context.Context, cmd *cobra.Command, _ []string) error {
	currentRollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Scope)
	if err != nil {
		return err
	}

	if _, err := c.Rig.Capsule().AbortRollout(ctx, &connect.Request[capsule.AbortRolloutRequest]{
		Msg: &capsule.AbortRolloutRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			RolloutId: currentRollout.GetRolloutId(),
			ProjectId: flags.GetProject(c.Scope),
		},
	}); err != nil {
		return err
	}

	cmd.Println("Current rollout aborted")

	return nil
}
