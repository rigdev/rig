package instance

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c *Cmd) restart(ctx context.Context, cmd *cobra.Command, args []string) error {
	arg := ""
	if len(args) > 1 {
		arg = args[1]
	}

	instanceID, err := c.provideInstanceID(ctx, capsule_cmd.CapsuleID, arg, cmd.ArgsLenAtDash())
	if err != nil {
		return err
	}

	if _, err := c.Rig.Capsule().RestartInstance(ctx, &connect.Request[capsule.RestartInstanceRequest]{
		Msg: &capsule.RestartInstanceRequest{
			CapsuleId:     capsule_cmd.CapsuleID,
			InstanceId:    instanceID,
			ProjectId:     c.Scope.GetCurrentContext().GetProject(),
			EnvironmentId: c.Scope.GetCurrentContext().GetEnvironment(),
		},
	}); err != nil {
		return err
	}

	cmd.Println("Instance restarted")

	return nil
}
