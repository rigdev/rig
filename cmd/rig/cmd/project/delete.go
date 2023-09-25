package project

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/spf13/cobra"
)

func (c Cmd) delete(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	req := &project.DeleteRequest{}

	_, err := c.Rig.Project().Delete(ctx, &connect.Request[project.DeleteRequest]{Msg: req})
	if err != nil {
		return err
	}

	cmd.Println("Successfully deleted project")
	return nil
}
