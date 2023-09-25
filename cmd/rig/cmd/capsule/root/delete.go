package root

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c Cmd) delete(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	if _, err := c.Rig.Capsule().Delete(ctx, &connect.Request[capsule.DeleteRequest]{
		Msg: &capsule.DeleteRequest{
			CapsuleId: capsule_cmd.CapsuleID,
		},
	}); err != nil {
		return err
	}

	cmd.Println("Delete capsule", capsule_cmd.CapsuleID)
	return nil
}
