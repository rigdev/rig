package environment

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/spf13/cobra"
)

func (c *Cmd) removeProject(ctx context.Context, cmd *cobra.Command, args []string) error {
	environmentID := args[0]
	projectID := args[1]

	req := &environment.UpdateRequest{
		EnvironmentId: environmentID,
		Updates: []*environment.Update{
			{
				Field: &environment.Update_RemoveProject{
					RemoveProject: projectID,
				},
			},
		},
	}

	_, err := c.Rig.Environment().Update(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}

	cmd.Printf("Successfully removed project %s from environment '%s'\n", projectID, environmentID)
	return nil
}
