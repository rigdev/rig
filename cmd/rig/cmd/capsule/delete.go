package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/spf13/cobra"
)

func delete(ctx context.Context, cmd *cobra.Command, nc rig.Client) error {
	if _, err := nc.Capsule().Delete(ctx, &connect.Request[capsule.DeleteRequest]{
		Msg: &capsule.DeleteRequest{
			CapsuleId: CapsuleID,
		},
	}); err != nil {
		return err
	}

	cmd.Println("Delete capsule", CapsuleID)
	return nil
}
