package user

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) lookup(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	u, id, err := common.GetUser(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	if outputJson {
		cmd.Println(common.ProtoToPrettyJson(u))
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows([]table.Row{
		{"ID", id},
		{"Name", u.GetProfile().GetFirstName() + " " + u.GetProfile().GetLastName()},
		{"Email", u.GetUserInfo().GetEmail()},
		{"Phone number", u.GetUserInfo().GetPhoneNumber()},
		{"Username", u.GetUserInfo().GetUsername()},
		{"First name", u.GetProfile().GetFirstName()},
		{"Last name", u.GetProfile().GetLastName()},
		{"Country", u.GetProfile().GetCountry()},
		{"Is Email verified", u.GetIsEmailVerified()},
		{"Is Phone verified", u.GetIsPhoneVerified()},
		{"Created at", u.GetUserInfo().GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")},
	})
	cmd.Println(t.Render())
	return nil
}
