package rollout

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) rollback(ctx context.Context, cmd *cobra.Command, args []string) error {
	rollout, err := c.getRollback(ctx, args)
	if err != nil {
		return err
	}

	req := connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Changes: []*capsule.Change{{
			Field: &capsule.Change_Rollback_{
				Rollback: &capsule.Change_Rollback{
					RollbackId: rollout,
				},
			},
		}},
		ProjectId:     c.Scope.GetCurrentContext().GetProject(),
		EnvironmentId: c.Scope.GetCurrentContext().GetEnvironment(),
	})

	resp, err := c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, req)
		} else {
			_, err = capsule_cmd.PromptAbortAndDeploy(ctx, c.Rig, c.Prompter, req)
		}
	}
	if err != nil {
		return err
	}
	cmd.Printf("rollback to %v initiated. New rollout has ID %v\n", rollout, resp.Msg.GetRolloutId())

	return nil
}

func (c *Cmd) getRollback(ctx context.Context, args []string) (uint64, error) {
	var rollout uint64
	var err error
	if len(args) > 1 {
		rollout, err = strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return 0, errors.InvalidArgumentErrorf("invalid rollout id - %v", err)
		}
	} else {
		resp, err := c.Rig.Capsule().ListRollouts(ctx, &connect.Request[capsule.ListRolloutsRequest]{
			Msg: &capsule.ListRolloutsRequest{
				CapsuleId:     capsule_cmd.CapsuleID,
				ProjectId:     c.Scope.GetCurrentContext().GetProject(),
				EnvironmentId: c.Scope.GetCurrentContext().GetEnvironment(),
				Pagination: &model.Pagination{
					Limit:      10,
					Descending: true,
				},
			},
		})
		if err != nil {
			return 0, err
		}

		if len(resp.Msg.GetRollouts()) == 0 {
			return 0, errors.NotFoundErrorf("no rollouts found")
		}

		rollouts := []string{}
		for _, r := range resp.Msg.GetRollouts() {
			rollouts = append(rollouts, strconv.FormatUint(r.GetRolloutId(), 10))
		}

		_, rolloutString, err := c.Prompter.Select("Select a rollout", rollouts)
		if err != nil {
			return 0, err
		}

		rollout, err = strconv.ParseUint(rolloutString, 10, 32)
		if err != nil {
			return 0, errors.InvalidArgumentErrorf("invalid rollout id - %v", err)
		}
	}

	return rollout, nil
}
