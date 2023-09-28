package rollout

import (
	"context"
	"errors"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c Cmd) rollback(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx

	rolloutID, err := c.getRollback(ctx)
	if err != nil {
		return err
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

func (c Cmd) getRollback(ctx context.Context) (uint64, error) {
	if rolloutID >= 0 {
		return uint64(rolloutID), nil
	}

	resp, err := c.Rig.Capsule().ListRollouts(ctx, connect.NewRequest(&capsule.ListRolloutsRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Pagination: &model.Pagination{
			Offset:     1,
			Descending: true,
		},
	}))
	if err != nil {
		return 0, err
	}

	for _, r := range resp.Msg.GetRollouts() {
		s := r.GetStatus().GetState()
		if s == capsule.RolloutState_ROLLOUT_STATE_DONE {
			return r.RolloutId, nil
		}
	}

	return 0, errors.New("no previous successful rollout")
}
