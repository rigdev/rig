package environment

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/spf13/cobra"
)

func (c *Cmd) delete(ctx context.Context, cmd *cobra.Command, args []string) error {
	var environmentID string
	if len(args) > 0 {
		environmentID = args[0]
	}

	req := &environment.DeleteRequest{
		EnvironmentId: environmentID,
		// TODO: Use API when updated.
		// Force:         force,
	}

	_, err := c.Rig.Environment().Delete(ctx, &connect.Request[environment.DeleteRequest]{Msg: req})
	if err != nil {
		return err
	}

	cmd.Printf("Successfully deleted environment '%s'\n", environmentID)
	return nil
}
