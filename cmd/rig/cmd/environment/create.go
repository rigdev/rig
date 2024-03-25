package environment

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, _ *cobra.Command, args []string) error {
	var id string
	if len(args) > 0 {
		id = args[0]
	}
	var cluster string
	if len(args) > 1 {
		cluster = args[1]
	}
	return c.createEnvironment(ctx, id, cluster, &useEnvironment)
}

func (c *Cmd) createEnvironment(ctx context.Context, name string, cluster string, useNewEnvironment *bool) error {
	var err error
	if name == "" {
		name, err = common.PromptInput("Environment:", common.ValidateSystemNameOpt)
		if err != nil {
			return err
		}
	}

	if cluster == "" {
		cluster, err = common.PromptInput("Cluster:", common.ValidateSystemNameOpt)
		if err != nil {
			return err
		}
	}

	initializers := []*environment.Update{}

	if _, err := c.Rig.Environment().Create(ctx, &connect.Request[environment.CreateRequest]{
		Msg: &environment.CreateRequest{
			EnvironmentId:     name,
			ClusterId:         cluster,
			Initializers:      initializers,
			NamespaceTemplate: namespaceTemplate,
		},
	}); err != nil {
		return err
	}

	fmt.Printf("Successfully created environment %s in cluster %s\n", name, cluster)

	if useNewEnvironment == nil {
		ok, err := common.PromptConfirm("Would you like to use this environment now?", true)
		if err != nil {
			return err
		}
		useNewEnvironment = &ok
	}

	if *useNewEnvironment {
		c.Cfg.GetCurrentContext().EnvironmentID = name
		if err := c.Cfg.Save(); err != nil {
			return err
		}

		fmt.Println("Changed environment successfully!")
	}

	return nil
}
