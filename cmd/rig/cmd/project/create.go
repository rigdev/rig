package project

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, _ []string) error {
	if name == "" {
		var err error
		name, err = common.PromptInput("Project name:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	initializers := []*project.Update{
		{
			Field: &project.Update_Name{
				Name: name,
			},
		},
	}

	res, err := c.Rig.Project().Create(ctx, &connect.Request[project.CreateRequest]{
		Msg: &project.CreateRequest{
			Initializers: initializers,
			ProjectId:    name,
		},
	})
	if err != nil {
		return err
	}

	p := res.Msg.GetProject()
	cmd.Printf("Successfully created project %s with id %s \n", name, p.GetProjectId())

	if useProject {
		c.Cfg.GetCurrentContext().ProjectID = p.GetProjectId()
		if err := c.Cfg.Save(); err != nil {
			return err
		}

		cmd.Println("Changed project successfully!")
	}

	return nil
}
