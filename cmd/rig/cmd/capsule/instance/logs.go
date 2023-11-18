package instance

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *Cmd) logs(ctx context.Context, cmd *cobra.Command, args []string) error {
	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}

	instanceID, err := c.provideInstanceID(ctx, capsule_cmd.CapsuleID, arg, cmd.ArgsLenAtDash())
	if err != nil {
		return err
	}

	duration, err := time.ParseDuration(since)
	if err != nil {
		return err
	}

	s, err := c.Rig.Capsule().Logs(ctx, &connect.Request[capsule.LogsRequest]{
		Msg: &capsule.LogsRequest{
			CapsuleId:  capsule_cmd.CapsuleID,
			InstanceId: instanceID,
			Follow:     follow,
			Since:      durationpb.New(duration),
		},
	})
	if err != nil {
		return err
	}

	return capsule_cmd.PrintLogs(s)
}
