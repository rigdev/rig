package env

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func set(ctx context.Context, args []string, cmd *cobra.Command, rc rig.Client) error {
	if len(args) != 2 {
		return errors.InvalidArgumentErrorf("expected key and value arguments")
	}

	r, err := cmd_capsule.GetCurrentRollout(ctx, rc)
	if err != nil {
		return err
	}

	cs := r.GetConfig().GetContainerSettings()

	if cs.GetEnvironmentVariables() == nil {
		cs.EnvironmentVariables = make(map[string]string)
	}
	cs.EnvironmentVariables[args[0]] = args[1]

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
