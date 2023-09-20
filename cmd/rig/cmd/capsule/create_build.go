package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
)

func CapsuleCreateBuild(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, nc rig.Client, cfg *cmd_config.Config, dockerClient *client.Client) error {
	var err error

	imageRef := imageRefFromFlags()
	if image == "" {
		imageRef, err = promptForImage(ctx, dockerClient)
		if err != nil {
			return err
		}
	}

	buildID, err := createBuild(ctx, nc, capsuleID, dockerClient, imageRef)
	if err != nil {
		return err
	}

	if !deploy {
		return nil
	}

	if _, err := nc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsuleID,
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
