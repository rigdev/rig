package image

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) addImage(ctx context.Context, cmd *cobra.Command, _ []string) error {
	var err error

	imageRef := imageRefFromFlags()
	if imageID == "" {
		imageRef, err = c.promptForImage(ctx)
		if err != nil {
			return err
		}
	}

	imageID, err := c.createImageInner(ctx, capsule_cmd.CapsuleID, imageRef)
	if err != nil {
		return err
	}

	if !deploy {
		return nil
	}

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{{
				Field: &capsule.Change_ImageId{
					ImageId: imageID,
				},
			}},
			ProjectId:     flags.GetProject(c.Scope),
			EnvironmentId: flags.GetEnvironment(c.Scope),
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, req)
		} else {
			_, err = capsule_cmd.PromptAbortAndDeploy(ctx, c.Rig, c.Prompter, req)
		}
	}
	if err != nil {
		return err
	}

	cmd.Printf("Deployed build %v \n", imageID)

	return nil
}
