package group

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

type groupField int64

const (
	groupUndefined groupField = iota
	groupIDField
	groupSetMetaData
	groupDeleteMetaData
)

func (f groupField) String() string {
	switch f {
	case groupIDField:
		return "Group ID"
	case groupSetMetaData:
		return "Set Metadata"
	case groupDeleteMetaData:
		return "Delete Metadata"
	default:
		return "Undefined"
	}
}

func (c *Cmd) update(ctx context.Context, cmd *cobra.Command, args []string) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	g, uid, err := common.GetGroup(ctx, identifier, c.Rig, c.Prompter)
	if err != nil {
		return err
	}

	fields := []string{
		groupIDField.String(),
		groupSetMetaData.String(),
		groupDeleteMetaData.String(),
		"Done",
	}

	updates := []*group.Update{}
	for {
		i, res, err := c.Prompter.Select("Choose a field to update:", fields)
		if err != nil {
			return err
		}
		if res == "Done" {
			break
		}
		u, err := c.promptGroupUpdate(groupField(i+1), g)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if u != nil {
			updates = append(updates, u)
		}
	}

	_, err = c.Rig.Group().Update(ctx, &connect.Request[group.UpdateRequest]{
		Msg: &group.UpdateRequest{
			GroupId: uid,
			Updates: updates,
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("Group %s updated\n", g.GetGroupId())
	return nil
}

func (c *Cmd) promptGroupUpdate(f groupField, g *group.Group) (*group.Update, error) {
	switch f {
	case groupIDField:
		name, err := c.Prompter.Input("ID:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(g.GetGroupId()))
		if err != nil {
			return nil, err
		}

		if name != g.GetGroupId() {
			return &group.Update{
				Field: &group.Update_GroupId{
					GroupId: groupID,
				},
			}, nil
		}
	case groupSetMetaData:
		key, err := c.Prompter.Input("Key:", common.ValidateNonEmptyOpt)
		if err != nil {
			return nil, err
		}
		value, err := c.Prompter.Input("Value:", common.ValidateNonEmptyOpt)
		if err != nil {
			return nil, err
		}

		return &group.Update{
			Field: &group.Update_SetMetadata{
				SetMetadata: &model.Metadata{
					Key:   key,
					Value: []byte(value),
				},
			},
		}, nil

	case groupDeleteMetaData:
		key, err := c.Prompter.Input("Key:", common.ValidateNonEmptyOpt)
		if err != nil {
			return nil, err
		}

		return &group.Update{
			Field: &group.Update_DeleteMetadataKey{
				DeleteMetadataKey: key,
			},
		}, nil
	default:
		return nil, nil
	}
	return nil, nil
}
