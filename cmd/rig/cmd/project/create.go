package project

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) create(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
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
		res, err := c.Rig.Project().Use(ctx, &connect.Request[project.UseRequest]{
			Msg: &project.UseRequest{
				ProjectId: p.GetProjectId(),
			},
		})
		if err != nil {
			return err
		}

		c.Cfg.GetCurrentContext().Project.ProjectID = p.GetProjectId()
		c.Cfg.GetCurrentContext().Project.ProjectToken = res.Msg.GetProjectToken()
		if err := c.Cfg.Save(); err != nil {
			return err
		}

		cmd.Println("Changed project successfully!")
	}

	return nil
}
