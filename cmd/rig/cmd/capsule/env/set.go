package env

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) set(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.InvalidArgumentErrorf("expected key and value arguments")
	}

	cs := &capsule.ContainerSettings{}

	r, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if r.GetConfig().GetContainerSettings() != nil {
		cs = r.GetConfig().GetContainerSettings()
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
			ProjectId:     flags.GetProject(c.Cfg),
			EnvironmentId: flags.GetEnvironment(c.Cfg),
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)

	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, req)
		} else {
			_, err = capsule_cmd.PromptAbortAndDeploy(ctx, c.Rig, req)
		}
	}
	if err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Environment variable %s=%s set", args[0], args[1]))

	return nil
}
