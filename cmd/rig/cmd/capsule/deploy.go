package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func CapsuleDeploy(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, nc rig.Client) error {
	var err error
	if buildID == "" {
		buildID, err = common.PromptInput("Enter Build ID:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}
	if _, err := nc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsuleID,
			Changes: []*capsule.Change{{
				Field: &capsule.Change_BuildId{BuildId: buildID},
			}},
		},
	}); err != nil {
		return err
	}

	cmd.Printf("Deployed build %v \n", buildID)

	return nil
}
