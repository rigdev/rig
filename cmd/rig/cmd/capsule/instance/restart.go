package instance

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func restart(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	instanceID, err := provideInstanceID(ctx, nc, capsule_cmd.CapsuleID, arg)
	if err != nil {
		return err
	}

	if _, err := nc.Capsule().RestartInstance(ctx, &connect.Request[capsule.RestartInstanceRequest]{
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
