package env

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) remove(ctx context.Context, cmd *cobra.Command, args []string) error {
	var key string
	var err error
	if len(args) > 0 {
		key = args[0]
	} else {
		key, err = common.PromptInput("key:", nil)
		if err != nil {
			return err
		}
	}

	r, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil {
		return err
	}

	cs := r.GetConfig().GetContainerSettings()

	if cs.GetEnvironmentVariables() == nil {
		cmd.Println("No environment variables set")
	}

	delete(cs.GetEnvironmentVariables(), key)

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{
				{
					Field: &capsule.Change_ContainerSettings{
						ContainerSettings: cs,
					},
				},
			},
			ProjectId:     c.Cfg.GetProject(),
			EnvironmentId: base.Flags.Environment,
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)
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

	return nil
}
