package rollout

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) rollback(ctx context.Context, cmd *cobra.Command, _ []string) error {
	rolloutID, err := c.getRollback(ctx)
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
		ProjectId:     c.Cfg.GetProject(),
		EnvironmentId: base.GetEnvironment(c.Cfg),
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

func (c *Cmd) getRollback(ctx context.Context) (uint64, error) {
	if rolloutID >= 0 {
		return uint64(rolloutID), nil
	}

	resp, err := c.Rig.Capsule().ListRollouts(ctx, connect.NewRequest(&capsule.ListRolloutsRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Pagination: &model.Pagination{
			Offset:     1,
			Descending: true,
		},
		ProjectId:     c.Cfg.GetProject(),
		EnvironmentId: base.GetEnvironment(c.Cfg),
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
