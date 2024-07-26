package config

import (
	"context"
	"fmt"
	"os"
	"slices"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *CmdWScope) useEnvironment(ctx context.Context, cmd *cobra.Command, args []string) error {
	var environment string
	var err error
	if len(args) == 0 {
		environment, err = c.promptForEnvironment(ctx)
		if err != nil {
			return err
		}
	} else {
		environment, err = c.environmentFromArg(ctx, args[0])
		if err != nil {
			return err
		}
	}

	c.Scope.GetCurrentContext().EnvironmentID = environment
	if err := c.Scope.GetCfg().Save(); err != nil {
		return err
	}

	cmd.Println("Changed environment successfully!")

	return nil
}

func (c *CmdWScope) promptForEnvironment(ctx context.Context) (string, error) {
	res, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
	if err != nil {
		return "", err
	}

	var es []string
	for _, e := range res.Msg.GetEnvironments() {
		es = append(es, e.GetEnvironmentId())
	}

	for {
		i, _, err := c.Prompter.Select("Environment: ", es)
		if err != nil {
			return "", err
		}

		environment := res.Msg.GetEnvironments()[i]
		if c.Scope.GetCurrentContext().GetProject() != "" && !environment.GetGlobal() &&
			!slices.Contains(environment.GetActiveProjects(), c.Scope.GetCurrentContext().GetProject()) {
			selectNew, err := c.Prompter.Confirm(
				fmt.Sprintf(
					"Warning: project '%s' is not active in environment '%s'.\nDo you want to select a different one?",
					c.Scope.GetCurrentContext().GetProject(),
					environment.GetEnvironmentId(),
				),
				false)
			if err != nil {
				return "", err
			}

			if !selectNew {
				return environment.GetEnvironmentId(), nil
			}
		} else {
			return environment.GetEnvironmentId(), nil
		}
	}
}

func (c *CmdWScope) environmentFromArg(ctx context.Context, environmentArg string) (string, error) {
	res, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
	if err != nil {
		return "", err
	}

	for _, e := range res.Msg.GetEnvironments() {
		if e.GetEnvironmentId() != environmentArg {
			continue
		}

		if c.Scope.GetCurrentContext().GetProject() != "" &&
			!slices.Contains(e.GetActiveProjects(), c.Scope.GetCurrentContext().GetProject()) {
			cont, err := c.Prompter.Confirm(
				fmt.Sprintf(
					"Warning: project '%s' is not active in environment '%s'.\nDo you want to continue anyways?",
					c.Scope.GetCurrentContext().GetProject(),
					e.GetEnvironmentId(),
				),
				true)
			if err != nil {
				return "", err
			}

			if !cont {
				os.Exit(0)
			}
		}

		return e.GetEnvironmentId(), nil
	}

	return "", errors.NotFoundErrorf("environment '%v' not found", environmentArg)
}
