package capsule

import (
	"context"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func CapsuleCreateBuild(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, nc rig.Client, cfg *base.Config) error {
	var err error
	if image == "" {
		image, err = utils.PromptGetInput("Enter image", utils.ValidateImage)
		if err != nil {
			return err
		}
	}

	// Generate a new tag for the build.
	buildID := strings.ReplaceAll(uuid.New().String()[:24], "-", "")

	if _, err := nc.Capsule().CreateBuild(ctx, &connect.Request[capsule.CreateBuildRequest]{
		Msg: &capsule.CreateBuildRequest{
			CapsuleId: capsuleID.String(),
			BuildId:   buildID,
			Image:     image,
		},
	}); err != nil {
		return err
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
