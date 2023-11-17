package project

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

type (
	projectField int32
)

const (
	projectUndefined projectField = iota
	projectName
)

func (p projectField) String() string {
	switch p {
	case projectName:
		return "name"
	default:
		return "undefined"
	}
}

func (c Cmd) update(ctx context.Context, cmd *cobra.Command, args []string) error {
	resp, err := c.Rig.Project().Get(ctx, &connect.Request[project.GetRequest]{Msg: &project.GetRequest{}})
	if err != nil {
		return err
	}

	if value != "" && field != "" {
		u, err := parseUpdate()
		if err != nil {
			return err
		}

		_, err = c.Rig.Project().Update(ctx, &connect.Request[project.UpdateRequest]{
			Msg: &project.UpdateRequest{
				Updates: []*project.Update{u},
			},
		})
		if err != nil {
			return err
		}

		cmd.Printf("Successfully updated project")
		return nil
	}

	fields := []string{
		projectName.String(),
		"Done",
	}

	updates := []*project.Update{}
	for {
		i, res, err := common.PromptSelect("Choose a field to update:", fields)
		if err != nil {
			return err
		}
		if res == "Done" {
			break
		}
		u, err := promptProjectUpdate(projectField(i+1), resp.Msg.GetProject())
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if u != nil {
			updates = append(updates, u)
		}
	}

	_, err = c.Rig.Project().Update(ctx, connect.NewRequest(&project.UpdateRequest{
		Updates: updates,
	}))
	if err != nil {
		return err
	}

	cmd.Println("Updated project")
	return nil
}

func promptProjectUpdate(f projectField, p *project.Project) (*project.Update, error) {
	switch f {
	case projectName:
		return &project.Update{
			Field: &project.Update_Name{
				Name: p.GetName(),
			},
		}, nil
	default:
		return nil, nil
	}
}

func parseUpdate() (*project.Update, error) {
	switch field {
	case common.FormatField(projectName.String()):
		return &project.Update{
			Field: &project.Update_Name{
				Name: value,
			},
		}, nil
	default:
		return nil, errors.InvalidArgumentErrorf("field %s is not supported", field)
	}
}
