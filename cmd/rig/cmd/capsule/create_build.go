package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
)

func CapsuleCreateBuild(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, nc rig.Client, cfg *cmd_config.Config) error {
	var err error
	if image == "" {
		image, err = common.PromptGetInput("Enter image", common.ValidateImageOpt)
		if err != nil {
			return err
		}
	}

	var buildID string
	if res, err := nc.Capsule().CreateBuild(ctx, &connect.Request[capsule.CreateBuildRequest]{
		Msg: &capsule.CreateBuildRequest{
			CapsuleId: capsuleID.String(),
			Image:     image,
		},
	}); err != nil {
		return err
	} else {
		buildID = res.Msg.GetBuildId()
	}

	if deploy {
		if _, err := nc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
			Msg: &capsule.DeployRequest{
				CapsuleId: capsuleID.String(),
				Changes: []*capsule.Change{{
					Field: &capsule.Change_BuildId{
						BuildId: buildID,
					},
				}},
			},
		}); err != nil {
			return err
		}

		cmd.Printf("Deployed build %v \n", buildID)
	} else {
		cmd.Printf("Image available as build %v\n", buildID)
	}

	return nil
}
