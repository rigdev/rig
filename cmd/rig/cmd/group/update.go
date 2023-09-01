package group

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

type groupField int64

const (
	groupUndefined groupField = iota
	groupName
	groupSetMetaData
	groupDeleteMetaData
)

func (f groupField) String() string {
	switch f {
	case groupName:
		return "Name"
	case groupSetMetaData:
		return "Set Metadata"
	case groupDeleteMetaData:
		return "Delete Metadata"
	default:
		return "Undefined"
	}
}

func GroupUpdate(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	g, uid, err := utils.GetGroup(ctx, identifier, nc)
	if err != nil {
		return err
	}

	fields := []string{
		groupName.String(),
		groupSetMetaData.String(),
		groupDeleteMetaData.String(),
		"Done",
	}

	updates := []*group.Update{}
	for {
		i, res, err := utils.PromptSelect("Choose a field to update:", fields, true)
		if err != nil {
			return err
		}
		if res == "Done" {
			break
		}
		u, err := promptGroupUpdate(groupField(i+1), g)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if u != nil {
			updates = append(updates, u)
		}
	}

	_, err = nc.Group().Update(ctx, &connect.Request[group.UpdateRequest]{
		Msg: &group.UpdateRequest{
			GroupId: uid,
			Updates: updates,
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("Group %s updated\n", g.GetName())
	return nil
}

func promptGroupUpdate(f groupField, g *group.Group) (*group.Update, error) {
	switch f {
	case groupName:
		name, err := utils.PromptGetInputWithDefault("Name:", utils.ValidateNonEmpty, g.GetName())
		if err != nil {
			return nil, err
		}

		if name != g.GetName() {
			return &group.Update{
				Field: &group.Update_Name{
					Name: name,
				},
			}, nil
		}
	case groupSetMetaData:
		key, err := utils.PromptGetInput("Key:", utils.ValidateNonEmpty)
		if err != nil {
			return nil, err
		}
		value, err := utils.PromptGetInput("Value:", utils.ValidateNonEmpty)
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
		key, err := utils.PromptGetInput("Key:", utils.ValidateNonEmpty)
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
