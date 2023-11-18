package instance

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c *Cmd) restart(ctx context.Context, cmd *cobra.Command, args []string) error {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	instanceID, err := c.provideInstanceID(ctx, capsule_cmd.CapsuleID, arg, cmd.ArgsLenAtDash())
	if err != nil {
		return err
	}

	if _, err := c.Rig.Capsule().RestartInstance(ctx, &connect.Request[capsule.RestartInstanceRequest]{
		Msg: &capsule.RestartInstanceRequest{
			CapsuleId:  capsule_cmd.CapsuleID,
			InstanceId: instanceID,
		},
	}); err != nil {
		return err
	}

	cmd.Println("Instance restarted")

	return nil
}
