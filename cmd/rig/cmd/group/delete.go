package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func GroupDelete(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}

	g, uid, err := common.GetGroup(ctx, identifier, nc)
	if err != nil {
		return err
	}

	_, err = nc.Group().Delete(ctx, &connect.Request[group.DeleteRequest]{
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
