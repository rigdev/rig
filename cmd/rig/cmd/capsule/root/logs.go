package root

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *Cmd) logs(ctx context.Context, _ *cobra.Command, _ []string) error {
	request := &capsule.LogsRequest{
		CapsuleId:          capsule_cmd.CapsuleID,
		Follow:             follow,
		ProjectId:          flags.GetProject(c.Scope),
		EnvironmentId:      flags.GetEnvironment(c.Scope),
		PreviousContainers: previousContainers,
	}

	if since != "" {
		duration, err := time.ParseDuration(since)
		if err != nil {
			return err
		}
		request.Since = durationpb.New(duration)
	}

	stream, err := c.Rig.Capsule().Logs(ctx, connect.NewRequest(request))
	if err != nil {
		return err
	}

	return capsule_cmd.PrintLogs(stream)
}
