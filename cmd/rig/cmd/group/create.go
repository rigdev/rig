package group

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, _ []string) error {
	var err error
	if groupID == "" {
		groupID, err = c.Prompter.Input("Group ID:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	updates := []*group.Update{
		{
			Field: &group.Update_GroupId{
				GroupId: groupID,
			},
		},
	}

	res, err := c.Rig.Group().Create(ctx, &connect.Request[group.CreateRequest]{
		Msg: &group.CreateRequest{
			Initializers: updates,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Created group:", groupID, res.Msg.GetGroup().GetGroupId())
	return nil
}
