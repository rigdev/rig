package builddeploy

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c Cmd) createBuild(cmd *cobra.Command, args []string) error {
	var err error
	ctx := c.Ctx

	imageRef := imageRefFromFlags()
	if !CmdFlags.image.IsSet(cmd) {
		imageRef, err = c.promptForImage(ctx)
		if err != nil {
			return err
		}
	}

	buildID, err := c.createBuildInner(ctx, cmd_capsule.CapsuleID, imageRef)
	if err != nil {
		return err
	}

	if !deploy {
		return nil
	}

	if _, err := c.Rig.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
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
