package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/spf13/cobra"
)

func ProjectDelete(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	req := &project.DeleteRequest{}

	_, err := nc.Project().Delete(ctx, &connect.Request[project.DeleteRequest]{Msg: req})
	if err != nil {
		return err
	}

	cmd.Println("Successfully deleted project")
	return nil
}
