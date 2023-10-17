package env

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) remove(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	var key string
	var err error
	if len(args) > 0 {
		key = args[0]
	} else {
		key, err = common.PromptInput("key:", nil)
		if err != nil {
			return err
		}
	}

	r, err := cmd_capsule.GetCurrentRollout(ctx, c.Rig)
	if err != nil {
		return err
	}

	cs := r.GetConfig().GetContainerSettings()

	if cs.GetEnvironmentVariables() == nil {
		cmd.Println("No environment variables set")
	}

	delete(cs.GetEnvironmentVariables(), key)

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: cmd_capsule.CapsuleID,
			Changes: []*capsule.Change{
				{
					Field: &capsule.Change_ContainerSettings{
						ContainerSettings: cs,
					},
				},
			},
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		_, err = cmd_capsule.AbortAndDeploy(ctx, cmd_capsule.CapsuleID, c.Rig, req)
	}
	if err != nil {
		return err
	}

	return nil
}
