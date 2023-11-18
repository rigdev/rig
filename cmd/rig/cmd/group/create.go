package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error
	if name == "" {
		name, err = common.PromptInput("Name:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	updates := []*group.Update{
		{
			Field: &group.Update_Name{
				Name: name,
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

	cmd.Println("Created group:", name, res.Msg.GetGroup().GetGroupId())
	return nil
}
