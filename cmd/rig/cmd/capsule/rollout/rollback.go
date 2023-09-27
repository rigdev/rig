package rollout

import (
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) rollback(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	if rolloutID == -1 {
		return errors.New("must supply a rollout-id")
	}

	if rolloutID < 0 {
		return errors.New("rollout-id cannot be negative")
	}

	resp, err := c.Rig.Capsule().Rollback(ctx, connect.NewRequest(&capsule.RollbackRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		RolloutId: uint64(rolloutID),
	}))
	if err != nil {
		return err
	}

	fmt.Printf("rollback to %v initiated. New rollout has ID %v\n", rolloutID, resp.Msg.GetRolloutId())

	return nil
}
