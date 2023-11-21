package root

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *Cmd) logs(ctx context.Context, _ *cobra.Command, _ []string) error {
	duration, err := time.ParseDuration(since)
	if err != nil {
		return err
	}

	stream, err := c.Rig.Capsule().Logs(ctx, connect.NewRequest(&capsule.LogsRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Follow:    follow,
		Since:     durationpb.New(duration),
	}))
	if err != nil {
		return err
	}

	return capsule_cmd.PrintLogs(stream)
}
