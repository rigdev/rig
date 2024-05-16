package environment

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
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

func (c *Cmd) createEnvironment(ctx context.Context,
	name string,
	clusterName string,
	useNewEnvironment *bool) error {
	var err error
	if name == "" {
		if !c.Scope.IsInteractive() {
			return fmt.Errorf("missing environment argument")
		}
		resp, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
		if err != nil {
			return err
		}
		var names []string
		for _, e := range resp.Msg.GetEnvironments() {
			names = append(names, e.GetEnvironmentId())
		}
		name, err = c.Prompter.Input(
			"Environment:", common.ValidateSystemNameOpt, common.ValidateUniqueOpt(names, "the environment name already exists"),
		)
		if err != nil {
			return err
		}
	}

	if clusterName == "" {
		if !c.Scope.IsInteractive() {
			return fmt.Errorf("missing cluster argument")
		}

		var names []string
		clusters, err := c.Rig.Cluster().List(ctx, connect.NewRequest(&cluster.ListRequest{}))
		if err != nil {
			return err
		}
		for _, c := range clusters.Msg.GetClusters() {
			names = append(names, c.GetClusterId())
		}

		_, clusterName, err = c.Prompter.Select("Cluster:", names, common.SelectEnableFilterOpt)
		if err != nil {
			return err
		}
	}

	initializers := []*environment.Update{
		{
			Field: &environment.Update_SetGlobal{
				SetGlobal: global,
			},
		},
	}

	for _, project := range addProjects {
		initializers = append(initializers, &environment.Update{
			Field: &environment.Update_AddProject{
				AddProject: project,
			},
		})
	}

	_, err = c.Rig.Environment().Create(ctx, &connect.Request[environment.CreateRequest]{
		Msg: &environment.CreateRequest{
			EnvironmentId:     name,
			ClusterId:         clusterName,
			Initializers:      initializers,
			NamespaceTemplate: namespaceTemplate,
			Ephemeral:         ephemeral,
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
		fmt.Printf("Successfully created environment %s in cluster %s\n", name, clusterName)
	}

	if useNewEnvironment == nil {
		if !c.Scope.IsInteractive() {
			return nil
		}
		ok, err := c.Prompter.Confirm("Would you like to use this environment now?", true)
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
