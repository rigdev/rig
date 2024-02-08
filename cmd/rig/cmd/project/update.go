package project

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
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

func (c *Cmd) update(ctx context.Context, cmd *cobra.Command, args []string) error {
	var projectID string
	if len(args) > 0 {
		projectID = args[0]
	} else {
		projectID = flags.GetProject(c.Cfg)
	}
	fmt.Println("projectID: ", projectID)

	resp, err := c.Rig.Project().Get(ctx, &connect.Request[project.GetRequest]{Msg: &project.GetRequest{
		ProjectId: projectID,
	}})
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
				ProjectId: projectID,
				Updates:   []*project.Update{u},
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
			continue
		}
		if u != nil {
			updates = append(updates, u)
		}
	}

	if len(updates) == 0 {
		cmd.Println("No updates to make")
		return nil
	}

	_, err = c.Rig.Project().Update(ctx, connect.NewRequest(&project.UpdateRequest{
		ProjectId: projectID,
		Updates:   updates,
	}))
	if err != nil {
		return err
	}

	cmd.Println("Updated project")
	return nil
}

func promptProjectUpdate(f projectField, p *project.Project) (*project.Update, error) {
	fmt.Println("f: ", f)
	switch f {
	case projectName:
		name, err := common.PromptInput("Name:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(p.GetName()))
		if err != nil {
			return nil, err
		}

		return &project.Update{
			Field: &project.Update_Name{
				Name: name,
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
