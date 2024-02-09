package rollout

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/capsule/rollout"
	"github.com/rigdev/rig-go-api/model"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) rollback(ctx context.Context, cmd *cobra.Command, args []string) error {
	rolloutID, err := c.getRollback(ctx, args[0])
	if err != nil {
		return err
	}

	req := connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Changes: []*capsule.Change{{
			Field: &capsule.Change_Rollback_{
				Rollback: &capsule.Change_Rollback{
					RollbackId: rolloutID,
				},
			},
		}},
		ProjectId:     flags.GetProject(c.Cfg),
		EnvironmentId: flags.GetEnvironment(c.Cfg),
	})

	resp, err := c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, c.Cfg, capsule_cmd.CapsuleID, req)
		} else {
			_, err = capsule_cmd.PromptAbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, c.Cfg, req)
		}
	}
	if err != nil {
		return err
	}
	cmd.Printf("rollback to %v initiated. New rollout has ID %v\n", rolloutID, resp.Msg.GetRolloutId())

	return nil
}

func (c *Cmd) getRollback(ctx context.Context, rolloutID string) (uint64, error) {
	if rolloutID != "" {
		// parse rolloutID to uint64
		rolloutID, err := strconv.ParseUint(rolloutID, 10, 64)
		if err != nil {
			return 0, errors.InvalidArgumentErrorf("invalid rollout ID: %v", rolloutID)
		}
		return rolloutID, nil
	}

	resp, err := c.Rig.Capsule().ListRollouts(ctx, connect.NewRequest(&capsule.ListRolloutsRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Pagination: &model.Pagination{
			Offset:     1,
			Descending: true,
		},
		ProjectId:     flags.GetProject(c.Cfg),
		EnvironmentId: flags.GetEnvironment(c.Cfg),
	}))
	if err != nil {
		return 0, err
	}

	for _, r := range resp.Msg.GetRollouts() {
		s := r.GetStatus().GetState()
		if s == rollout.State_STATE_STOPPED {
			return r.RolloutId, nil
		}
	}

	return 0, errors.New("no previous successful rollout")
}
