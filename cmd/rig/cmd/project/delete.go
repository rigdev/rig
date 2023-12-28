package project

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/spf13/cobra"
)

func (c *Cmd) delete(ctx context.Context, cmd *cobra.Command, _ []string) error {
	req := &project.DeleteRequest{}

	_, err := c.Rig.Project().Delete(ctx, &connect.Request[project.DeleteRequest]{Msg: req})
	if err != nil {
		return err
	}

	cmd.Println("Successfully deleted project")
	return nil
}
