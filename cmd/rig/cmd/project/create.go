package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func ProjectCreate(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client, cfg *base.Config) error {
	if name == "" {
		var err error
		name, err = common.PromptGetInput("Project name:", common.ValidateNonEmpty)
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

	res, err := nc.Project().Create(ctx, &connect.Request[project.CreateRequest]{
		Msg: &project.CreateRequest{
			Initializers: initializers,
		},
	})
	if err != nil {
		return err
	}

	p := res.Msg.GetProject()
	cmd.Printf("Successfully created project %s with id %s \n", name, p.GetProjectId())

	if useProject {
		res, err := nc.Project().Use(ctx, &connect.Request[project.UseRequest]{
			Msg: &project.UseRequest{
				ProjectId: p.GetProjectId(),
			},
		})
		if err != nil {
			return err
		}

		cfg.Context().Project.ProjectID = uuid.UUID(p.GetProjectId())
		cfg.Context().Project.ProjectToken = res.Msg.GetProjectToken()
		if err := cfg.Save(); err != nil {
			return err
		}

		cmd.Println("Changed project successfully!")
	}

	return nil
}
