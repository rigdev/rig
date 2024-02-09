package root

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) cmdArgs(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) == 0 && !deleteCmd {
		return errors.InvalidArgumentErrorf("must supply capsule cmd and arguments as arguments")
	}
	if len(args) > 0 && deleteCmd {
		return errors.InvalidArgumentErrorf("cannot both supply arguments and --delete")
	}

	containerSettings := &capsule.ContainerSettings{}

	r, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if r.GetConfig().GetContainerSettings() != nil {
		containerSettings = r.GetConfig().GetContainerSettings()
	}

	if deleteCmd {
		containerSettings.Command = ""
		containerSettings.Args = nil
	} else {
		containerSettings.Command = args[0]
		containerSettings.Args = args[1:]
	}

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{{
				Field: &capsule.Change_ContainerSettings{
					ContainerSettings: containerSettings,
				},
			}},
			ProjectId:     flags.GetProject(c.Cfg),
			EnvironmentId: flags.GetEnvironment(c.Cfg),
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

	cmd.Println("Capsule Command updated")

	return nil
}
