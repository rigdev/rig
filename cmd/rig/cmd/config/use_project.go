package config

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *CmdWScope) useProject(ctx context.Context, cmd *cobra.Command, args []string) error {
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

	envs, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{
		ProjectFilter: projectID,
	}))
	if err != nil {
		return err
	}

	for _, e := range envs.Msg.GetEnvironments() {
		if e.GetEnvironmentId() == c.Scope.GetCurrentContext().GetEnvironment() {
			return nil
		}
	}

	chooseEnv, err := c.Prompter.Confirm(
		"The project is not active in the current environment. Would you like to select a new environment as well?",
		false)
	if err != nil {
		return err
	}

	if !chooseEnv {
		return nil
	}

	env, err := c.promptForEnvironment(ctx)
	if err != nil {
		return err
	}

	c.Scope.GetCurrentContext().EnvironmentID = env
	if err := c.Scope.GetCfg().Save(); err != nil {
		return err
	}

	cmd.Println("Changed environment successfully!")
	return nil
}

func (c *CmdWScope) promptForProjectID(ctx context.Context) (string, error) {
	res, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return "", err
	}

	var ps []string
	for _, p := range res.Msg.GetProjects() {
		ps = append(ps, p.GetProjectId())
	}

	i, _, err := c.Prompter.Select("Project: ", ps)
	if err != nil {
		return "", err
	}

	project := res.Msg.GetProjects()[i]

	return project.GetProjectId(), nil
}

func (c *CmdWScope) projectIDFromArg(ctx context.Context, projectArg string) (string, error) {
	if projectArg != "" {
		return projectArg, nil
	}

	res, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return "", err
	}

	for _, p := range res.Msg.GetProjects() {
		if p.GetProjectId() != projectArg {
			continue
		}
		return p.GetProjectId(), nil
	}

	return "", errors.NotFoundErrorf("project '%v' not found", projectArg)
}
