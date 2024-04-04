package project

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) delete(ctx context.Context, cmd *cobra.Command, args []string) error {
	var projectID string
	if len(args) > 0 {
		projectID = args[0]
	} else {
		projectID = flags.GetProject(c.Scope)
	}

	req := &project.DeleteRequest{
		ProjectId: projectID,
	}

	_, err := c.Rig.Project().Delete(ctx, &connect.Request[project.DeleteRequest]{Msg: req})
	if err != nil {
		return err
	}

	cmd.Println("Successfully deleted project")
	return nil
}
