package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func config(ctx context.Context, cmd *cobra.Command, rc rig.Client) error {
	if len(args) > 0 && command == "" {
		return errors.InvalidArgumentErrorf("command must be set when args are provided")
	}

	var cs []*capsule.Change

	if command != "" {
		r, err := GetCurrentRollout(ctx, rc)
		if err != nil {
			return err
		}
		containerSettings := r.GetConfig().GetContainerSettings()
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

	if _, err := rc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: CapsuleID,
			Changes:   cs,
		},
	}); err != nil {
		return err
	}

	cmd.Println("Capsule configuration updated")

	return nil
}
