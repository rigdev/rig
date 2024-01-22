package user

import (
	"context"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) getSettings(ctx context.Context, cmd *cobra.Command, _ []string) error {
	res, err := c.Rig.UserSettings().GetSettings(ctx, &connect.Request[settings.GetSettingsRequest]{})
	if err != nil {
		return err
	}
	settings := res.Msg.GetSettings()

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(settings, flags.Flags.OutputType)
	}

	rowsLogin := []table.Row{}
	for i, l := range settings.GetLoginMechanisms() {
		if i == 0 {
			rowsLogin = append(rowsLogin, table.Row{"Login Mechanisms", l})
			continue
		}
		rowsLogin = append(rowsLogin, table.Row{"", l})
	}
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRow(table.Row{"Allow Register", settings.GetAllowRegister()})
	t.AppendRow(table.Row{"Verify Email Required", settings.GetIsVerifiedEmailRequired()})
	t.AppendRow(table.Row{"Verify Phone Required", settings.GetIsVerifiedPhoneRequired()})
	t.AppendRow(table.Row{"Access Token TTL", settings.GetAccessTokenTtl().AsDuration()})
	t.AppendRow(table.Row{"Refresh Token TTL", settings.GetRefreshTokenTtl().AsDuration()})
	t.AppendRow(table.Row{"Verification Code TTL", settings.GetVerificationCodeTtl().AsDuration()})
	t.AppendRow(table.Row{"Password hashing", settings.GetPasswordHashing().GetMethod()})
	for _, r := range rowsLogin {
		t.AppendRow(r)
	}
	cmd.Println(t.Render())

	return nil
}
