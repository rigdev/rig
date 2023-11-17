package root

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) config(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) > 0 && command == "" {
		return errors.InvalidArgumentErrorf("command must be set when args are provided")
	}

	var cs []*capsule.Change

	if command != "" {
		r, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig)
		if err != nil {
			return err
		}
		containerSettings := r.GetConfig().GetContainerSettings()
		if containerSettings == nil {
			containerSettings = &capsule.ContainerSettings{}
		}
		containerSettings.Command = command

		if len(args) > 0 {
			containerSettings.Args = args
		}

		cs = append(cs, &capsule.Change{
			Field: &capsule.Change_ContainerSettings{ContainerSettings: containerSettings},
		})
	}

	if cmd.Flags().Changed("auto-add-service-account") {
		autoAdd, err := cmd.Flags().GetBool("auto-add-service-account")
		if err != nil {
			return err
		}

		cs = append(cs, &capsule.Change{
			Field: &capsule.Change_AutoAddRigServiceAccounts{AutoAddRigServiceAccounts: autoAdd},
		})
	}

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes:   cs,
		},
	}

	_, err := c.Rig.Capsule().Deploy(ctx, req)
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

	cmd.Println("Capsule configuration updated")

	return nil
}
