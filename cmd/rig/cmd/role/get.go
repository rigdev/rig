package role

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/role"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, args []string) error {
	roleID := ""
	if len(args) > 0 {
		roleID = args[0]
	} else {
		resp, err := c.Rig.Role().List(ctx, connect.NewRequest(&role.ListRequest{}))
		if err != nil {
			return err
		}

		var roleIDs []string
		for _, r := range resp.Msg.GetRoles() {
			roleIDs = append(roleIDs, r.GetRoleId())
		}

		_, roleID, err = c.Prompter.Select("Select role to get", roleIDs)
		if err != nil {
			return err
		}
	}

	resp, err := c.Rig.Role().Get(ctx, connect.NewRequest(&role.GetRequest{
		RoleId: roleID,
	}))
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetRole(), flags.Flags.OutputType)
	}

	roleType := resp.Msg.GetRole().GetRoleId()
	if v, ok := resp.Msg.GetRole().GetMetadata()["roleType"]; ok {
		roleType = string(v)
	}
	cmd.Println(fmt.Sprintf("Role ID: %s\nRole Type: %s\nProject: %s\nEnvironment: %s", resp.Msg.GetRole().GetRoleId(),
		roleType,
		resp.Msg.GetRole().GetPermissions()[0].GetScope().GetProject(),
		resp.Msg.GetRole().GetPermissions()[0].GetScope().GetEnvironment()),
	)
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Permissions", "Action", "Resource"})
	for i, p := range resp.Msg.GetRole().GetPermissions() {
		t.AppendRow(table.Row{i + 1, p.GetAction(), p.GetScope().GetResource()})
	}
	cmd.Println(t.Render())

	return nil
}
