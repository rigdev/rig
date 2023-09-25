package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func (c Cmd) use(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	var projectID uuid.UUID
	var err error
	if len(args) == 0 {
		projectID, err = c.promptForProjectID(ctx)
	} else {
		projectID, err = c.projectIDFromArg(ctx, args[0])
	}

	res, err := c.Rig.Project().Use(ctx, &connect.Request[project.UseRequest]{
		Msg: &project.UseRequest{
			ProjectId: projectID.String(),
		},
	})
	if err != nil {
		return err
	}

	c.Cfg.GetCurrentContext().Project.ProjectID = projectID
	c.Cfg.GetCurrentContext().Project.ProjectToken = res.Msg.GetProjectToken()
	if err := c.Cfg.Save(); err != nil {
		return err
	}

	cmd.Println("Changed project successfully!")

	return nil
}

func (c Cmd) promptForProjectID(ctx context.Context) (uuid.UUID, error) {
	res, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return uuid.Nil, err
	}

	var ps []string
	for _, p := range res.Msg.GetProjects() {
		ps = append(ps, p.GetName())
	}

	i, _, err := common.PromptSelect("Project: ", ps)
	if err != nil {
		return uuid.Nil, err
	}

	projectID, err := uuid.Parse(res.Msg.GetProjects()[i].GetProjectId())
	if err != nil {
		return uuid.Nil, err
	}

	return projectID, nil
}

func (c Cmd) projectIDFromArg(ctx context.Context, projectArg string) (uuid.UUID, error) {
	if id, err := uuid.Parse(projectArg); err == nil {
		return id, nil
	}
	res, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return uuid.Nil, err
	}

	for _, p := range res.Msg.GetProjects() {
		if p.GetName() != projectArg {
			continue
		}
		projectID, err := uuid.Parse(p.GetProjectId())
		if err != nil {
			return uuid.Nil, err
		}
		return projectID, nil
	}

	return uuid.Nil, errors.NotFoundErrorf("project '%v' not found", projectArg)
}
