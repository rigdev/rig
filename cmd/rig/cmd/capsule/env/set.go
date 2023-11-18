package env

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) set(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.InvalidArgumentErrorf("expected key and value arguments")
	}

	r, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig)
	if err != nil {
		return err
	}

	cs := r.GetConfig().GetContainerSettings()

	if cs == nil {
		cs = &capsule.ContainerSettings{}
	}

	if cs.GetEnvironmentVariables() == nil {
		cs.EnvironmentVariables = make(map[string]string)
	}
	cs.EnvironmentVariables[args[0]] = args[1]

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
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)

	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, capsule_cmd.CapsuleID, req)
		} else {
			_, err = capsule_cmd.PromptAbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, req)

		}
	}
	if err != nil {
		return err
	}

	return nil

}
