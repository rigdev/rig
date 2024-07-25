package instance

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *Cmd) logs(ctx context.Context, cmd *cobra.Command, args []string) error {
	arg := ""
	if len(args) > 1 {
		arg = args[1]
	}

	instanceID, err := c.provideInstanceID(ctx, capsule_cmd.CapsuleID, arg, cmd.ArgsLenAtDash())
	if err != nil {
		return err
	}

	request := &connect.Request[capsule.LogsRequest]{
		Msg: &capsule.LogsRequest{
			CapsuleId:          capsule_cmd.CapsuleID,
			InstanceId:         instanceID,
			Follow:             follow,
			ProjectId:          c.Scope.GetCurrentContext().GetProject(),
			EnvironmentId:      c.Scope.GetCurrentContext().GetEnvironment(),
			PreviousContainers: previousContainers,
		},
	}

	if since != "" {
		duration, err := time.ParseDuration(since)
		if err != nil {
			return err
		}

		request.Msg.Since = durationpb.New(duration)
	}

	s, err := c.Rig.Capsule().Logs(ctx, request)
	if err != nil {
		return err
	}

	return capsule_cmd.PrintLogs(s)
}
