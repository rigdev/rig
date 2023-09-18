package capsule

import (
	"context"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func CapsuleLogs(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, nc rig.Client) error {
	instanceID, err := provideInstanceID(ctx, nc, capsuleID, instanceID)
	if err != nil {
		return err
	}

	s, err := nc.Capsule().Logs(ctx, &connect.Request[capsule.LogsRequest]{
		Msg: &capsule.LogsRequest{
			CapsuleId:  capsuleID,
			InstanceId: instanceID,
			Follow:     follow,
		},
	})
	if err != nil {
		return err
	}

	for s.Receive() {
		switch v := s.Msg().GetLog().GetMessage().GetMessage().(type) {
		case *capsule.LogMessage_Stdout:
			os.Stdout.WriteString(s.Msg().GetLog().GetTimestamp().AsTime().Format(base.RFC3339NanoFixed))
			os.Stdout.WriteString(": ")
			if _, err := os.Stdout.Write(v.Stdout); err != nil {
				return err
			}
		case *capsule.LogMessage_Stderr:
			os.Stderr.WriteString(s.Msg().GetLog().GetTimestamp().AsTime().Format(base.RFC3339NanoFixed))
			os.Stderr.WriteString(": ")
			if _, err := os.Stderr.Write(v.Stderr); err != nil {
				return err
			}
		default:
			return errors.InvalidArgumentErrorf("invalid log message")
		}
	}

	return s.Err()
}
