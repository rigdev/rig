package builddeploy

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) createBuild(ctx context.Context, cmd *cobra.Command, _ []string) error {
	var err error

	imageRef := imageRefFromFlags()
	if image == "" {
		imageRef, err = c.promptForImage(ctx)
		if err != nil {
			return err
		}
	}

	buildID, err := c.createBuildInner(ctx, capsule_cmd.CapsuleID, imageRef)
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
				Field: &capsule.Change_BuildId{
					BuildId: buildID,
				},
			}},
			ProjectId:     flags.GetProject(c.Cfg),
			EnvironmentId: flags.GetEnvironment(c.Cfg),
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, req)
		} else {
			_, err = capsule_cmd.PromptAbortAndDeploy(ctx, c.Rig, req)
		}
	}
	if err != nil {
		return err
	}

	cmd.Printf("Deployed build %v \n", buildID)

	return nil
}
