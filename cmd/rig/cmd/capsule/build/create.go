package build

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
)

func create(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client, cfg *cmd_config.Config, dockerClient *client.Client) error {
	var err error

	imageRef := cmd_capsule.ImageRefFromFlags()
	if image == "" {
		imageRef, err = cmd_capsule.PromptForImage(ctx, dockerClient)
		if err != nil {
			return err
		}
	}

	buildID, err := cmd_capsule.CreateBuild(ctx, nc, cmd_capsule.CapsuleID, dockerClient, imageRef)
	if err != nil {
		return err
	}

	if !deploy {
		return nil
	}

	if _, err := nc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: cmd_capsule.CapsuleID,
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

	return nil
}
