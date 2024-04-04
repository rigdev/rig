package config

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) useProject(ctx context.Context, cmd *cobra.Command, args []string) error {
	var projectID string
	var err error
	if len(args) == 0 {
		projectID, err = c.promptForProjectID(ctx)
		if err != nil {
			return err
		}
	} else {
		projectID, err = c.projectIDFromArg(ctx, args[0])
		if err != nil {
			return err
		}
	}

	c.Scope.GetCurrentContext().ProjectID = projectID
	if err := c.Scope.GetCfg().Save(); err != nil {
		return err
	}

	cmd.Println("Changed project successfully!")

	return nil
}

func (c *Cmd) promptForProjectID(ctx context.Context) (string, error) {
	res, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return "", err
	}

	var ps []string
	for _, p := range res.Msg.GetProjects() {
		ps = append(ps, p.GetName())
	}

	i, _, err := common.PromptSelect("Project: ", ps)
	if err != nil {
		return "", err
	}

	projectID := res.Msg.GetProjects()[i].GetProjectId()
	return projectID, nil
}

func (c *Cmd) projectIDFromArg(ctx context.Context, projectArg string) (string, error) {
	if projectArg != "" {
		return projectArg, nil
	}

	res, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return "", err
	}

	for _, p := range res.Msg.GetProjects() {
		if p.GetName() != projectArg {
			continue
		}
		return p.GetProjectId(), nil
	}

	return "", errors.NotFoundErrorf("project '%v' not found", projectArg)
}
