package group

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) delete(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}

	g, uid, err := common.GetGroup(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	_, err = c.Rig.Group().Delete(ctx, &connect.Request[group.DeleteRequest]{
		Msg: &group.DeleteRequest{
			GroupId: uid,
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("Group %s deleted\n", g.GetName())
	return nil
}
