package config

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) useEnvironment(ctx context.Context, cmd *cobra.Command, args []string) error {
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

func (c *Cmd) promptForEnvironment(ctx context.Context) (string, error) {
	res, err := c.Rig.Environment().List(ctx, &connect.Request[environment.ListRequest]{})
	if err != nil {
		return "", err
	}

	var es []string
	for _, e := range res.Msg.GetEnvironments() {
		es = append(es, e.GetEnvironmentId())
	}

	i, _, err := common.PromptSelect("Environment: ", es)
	if err != nil {
		return "", err
	}

	environment := res.Msg.GetEnvironments()[i].GetEnvironmentId()
	return environment, nil
}

func (c *Cmd) environmentFromArg(ctx context.Context, environmentArg string) (string, error) {
	if environmentArg != "" {
		return environmentArg, nil
	}

	res, err := c.Rig.Environment().List(ctx, &connect.Request[environment.ListRequest]{})
	if err != nil {
		return "", err
	}

	for _, e := range res.Msg.GetEnvironments() {
		if e.GetEnvironmentId() == environmentArg {
			return e.GetEnvironmentId(), nil
		}
	}

	return "", errors.NotFoundErrorf("project '%v' not found", environmentArg)
}
