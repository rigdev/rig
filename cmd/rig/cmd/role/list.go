package role

import (
	"context"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/role"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) list(ctx context.Context, cmd *cobra.Command, _ []string) error {
	resp, err := c.Rig.Role().List(ctx, connect.NewRequest(&role.ListRequest{
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
	}))
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetRoles(), flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Roles", "ID", "Project", "Environment", "Role Type"})
	for i, r := range resp.Msg.GetRoles() {
		project := "*"
		environment := "*"
		if r.GetPermissions() != nil && len(r.GetPermissions()) > 0 {
			project = r.GetPermissions()[0].GetScope().GetProject()
			environment = r.GetPermissions()[0].GetScope().GetEnvironment()
		}

		roleType, ok := r.Metadata["roleType"]
		if !ok {
			roleType = []byte(r.GetRoleId())
		}

		t.AppendRow(table.Row{i + 1, r.GetRoleId(), project, environment, string(roleType)})
	}
	cmd.Println(t.Render())
	return nil
}
