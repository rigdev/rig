package root

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c *Cmd) delete(ctx context.Context, cmd *cobra.Command, _ []string) error {
	if _, err := c.Rig.Capsule().Delete(ctx, &connect.Request[capsule.DeleteRequest]{
		Msg: &capsule.DeleteRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			ProjectId: c.Cfg.GetProject(),
		},
	}); err != nil {
		return err
	}

	cmd.Println("Delete capsule", capsule_cmd.CapsuleID)
	return nil
}
