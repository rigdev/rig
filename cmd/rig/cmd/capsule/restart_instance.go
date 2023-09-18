package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/spf13/cobra"
)

func CapsuleRestartInstance(ctx context.Context, cmd *cobra.Command, capsuleID CapsuleID, nc rig.Client) error {
	instanceID, err := provideInstanceID(ctx, nc, capsuleID, instanceID)
	if err != nil {
		return err
	}

	if _, err := nc.Capsule().RestartInstance(ctx, &connect.Request[capsule.RestartInstanceRequest]{
		Msg: &capsule.RestartInstanceRequest{
			CapsuleId:  capsuleID,
			InstanceId: instanceID,
		},
	}); err != nil {
		return err
	}

	cmd.Println("Instance restarted")

	return nil
}
