package env

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func remove(ctx context.Context, args []string, cmd *cobra.Command, rc rig.Client) error {
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

	r, err := cmd_capsule.GetCurrentRollout(ctx, rc)
	if err != nil {
		return err
	}

	cs := r.GetConfig().GetContainerSettings()

	if cs.GetEnvironmentVariables() == nil {
		cmd.Println("No environment variables set")
	}

	delete(cs.GetEnvironmentVariables(), key)

	if _, err := rc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
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
	}); err != nil {
		return err
	}

	return nil
}
