package environment

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/spf13/cobra"
)

func (c *Cmd) update(ctx context.Context, cmd *cobra.Command, args []string) error {
	environmentID := args[0]

	var updates []*environment.Update

	for _, a := range addProjects {
		updates = append(updates, &environment.Update{
			Field: &environment.Update_AddProject{
				AddProject: a,
			},
		})
	}

	for _, r := range removeProjects {
		updates = append(updates, &environment.Update{
			Field: &environment.Update_RemoveProject{
				RemoveProject: r,
			},
		})
	}

	if cmd.Flags().Changed("set-global") {
		updates = append(updates, &environment.Update{
			Field: &environment.Update_SetGlobal{
				SetGlobal: updateGlobal,
			},
		})
	}

	req := &environment.UpdateRequest{
		EnvironmentId: environmentID,
		Updates:       updates,
	}

	_, err := c.Rig.Environment().Update(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}

	cmd.Printf("Successfully updated environment '%s'\n", environmentID)
	return nil
}
