package group

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) delete(ctx context.Context, cmd *cobra.Command, args []string) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}

	g, uid, err := common.GetGroup(ctx, identifier, c.Rig, c.Prompter)
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

	cmd.Printf("Group %s deleted\n", g.GetGroupId())
	return nil
}
