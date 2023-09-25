package env

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) set(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	if len(args) != 2 {
		return errors.InvalidArgumentErrorf("expected key and value arguments")
	}

	r, err := cmd_capsule.GetCurrentRollout(ctx, c.Rig)
	if err != nil {
		return err
	}

	cs := r.GetConfig().GetContainerSettings()

	if cs.GetEnvironmentVariables() == nil {
		cs.EnvironmentVariables = make(map[string]string)
	}
	cs.EnvironmentVariables[args[0]] = args[1]

	if _, err := c.Rig.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
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
