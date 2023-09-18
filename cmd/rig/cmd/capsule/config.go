package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/spf13/cobra"
)

func CapsuleConfig(ctx context.Context, cmd *cobra.Command, capsuleID CapsuleID, nc rig.Client) error {
	var cs []*capsule.Change

	if cmd.Flags().Changed("auto-add-service-account") {
		autoAdd, err := cmd.Flags().GetBool("auto-add-service-account")
		if err != nil {
			return err
		}

		cs = append(cs, &capsule.Change{
			Field: &capsule.Change_AutoAddRigServiceAccounts{AutoAddRigServiceAccounts: autoAdd},
		})
	}

	if _, err := nc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsuleID,
			Changes:   cs,
		},
	}); err != nil {
		return err
	}

	cmd.Println("Capsule configuration updated")

	return nil
}
