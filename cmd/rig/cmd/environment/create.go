package environment

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
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
		if !c.Scope.IsInteractive() {
			return fmt.Errorf("missing environment argument")
		}
		name, err = common.PromptInput("Environment:", common.ValidateSystemNameOpt)
		if err != nil {
			return err
		}
	}

	if cluster == "" {
		if !c.Scope.IsInteractive() {
			return fmt.Errorf("missing cluster argument")
		}
		cluster, err = common.PromptInput("Cluster:", common.ValidateSystemNameOpt)
		if err != nil {
			return err
		}
	}

	initializers := []*environment.Update{}

	_, err = c.Rig.Environment().Create(ctx, &connect.Request[environment.CreateRequest]{
		Msg: &environment.CreateRequest{
			EnvironmentId:     name,
			ClusterId:         cluster,
			Initializers:      initializers,
			NamespaceTemplate: namespaceTemplate,
		},
	})
	if errors.IsAlreadyExists(err) {
		if failIfExists {
			return err
		}

		fmt.Printf("Environment %s already exists\n", name)
	} else if err != nil {
		return err
	} else {
		fmt.Printf("Successfully created environment %s in cluster %s\n", name, cluster)
	}

	if useNewEnvironment == nil {
		if !c.Scope.IsInteractive() {
			return nil
		}
		ok, err := common.PromptConfirm("Would you like to use this environment now?", true)
		if err != nil {
			return err
		}
		useNewEnvironment = &ok
	}

	if *useNewEnvironment {
		c.Scope.GetCurrentContext().EnvironmentID = name
		if err := c.Scope.GetCfg().Save(); err != nil {
			return err
		}

		fmt.Println("Changed environment successfully!")
	}

	return nil
}
