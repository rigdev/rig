package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func ProjectUse(ctx context.Context, cmd *cobra.Command, args []string, client rig.Client, cfg *cmd_config.Config, logger *zap.Logger) error {
	var projectID uuid.UUID
	var err error
	if len(args) == 0 {
		projectID, err = promptForProjectID(ctx, client)
	} else {
		projectID, err = projectIDFromArg(ctx, client, args[0])
	}

	res, err := client.Project().Use(ctx, &connect.Request[project.UseRequest]{
		Msg: &project.UseRequest{
			ProjectId: projectID.String(),
		},
	})
	if err != nil {
		return err
	}

	cfg.GetCurrentContext().Project.ProjectID = projectID
	cfg.GetCurrentContext().Project.ProjectToken = res.Msg.GetProjectToken()
	if err := cfg.Save(); err != nil {
		return err
	}

	cmd.Println("Changed project successfully!")

	return nil
}

func promptForProjectID(ctx context.Context, client rig.Client) (uuid.UUID, error) {
	res, err := client.Project().List(ctx, &connect.Request[project.ListRequest]{})
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

func projectIDFromArg(ctx context.Context, client rig.Client, projectArg string) (uuid.UUID, error) {
	if id, err := uuid.Parse(projectArg); err == nil {
		return id, nil
	}
	res, err := client.Project().List(ctx, &connect.Request[project.ListRequest]{})
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
