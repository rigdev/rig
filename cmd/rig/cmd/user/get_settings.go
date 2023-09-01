package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func UserGetSettings(ctx context.Context, cmd *cobra.Command, nc rig.Client) error {
	res, err := nc.UserSettings().GetSettings(ctx, &connect.Request[settings.GetSettingsRequest]{})
	if err != nil {
		return err
	}
	settings := res.Msg.GetSettings()

	if outputJson {
		cmd.Println(utils.ProtoToPrettyJson(settings))
		return nil
	}

	rows_login := []table.Row{}
	for i, l := range settings.GetLoginMechanisms() {
		if i == 0 {
			rows_login = append(rows_login, table.Row{"Login Mechanisms", l})
			continue
		}
		rows_login = append(rows_login, table.Row{"", l})
	}

	oauthSettings := settings.GetOauthSettings()
	rows_oauth := []table.Row{}
	if oauthSettings.GetGoogle().GetAllowRegister() {
		if len(rows_oauth) == 0 {
			rows_oauth = append(rows_oauth, table.Row{"Oauth Providers", "Google"})
		} else {
			rows_oauth = append(rows_oauth, table.Row{"", "Google"})
		}
	}
	if oauthSettings.GetFacebook().GetAllowRegister() {
		if len(rows_oauth) == 0 {
			rows_oauth = append(rows_oauth, table.Row{"Oauth Providers", "Facebook"})
		} else {
			rows_oauth = append(rows_oauth, table.Row{"", "Facebook"})
		}
	}
	if oauthSettings.GetGithub().GetAllowRegister() {
		if len(rows_oauth) == 0 {
			rows_oauth = append(rows_oauth, table.Row{"Oauth Providers", "Github"})
		} else {
			rows_oauth = append(rows_oauth, table.Row{"", "Github"})
		}
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
	for _, r := range rows_login {
		t.AppendRow(r)
	}
	for _, r := range rows_oauth {
		t.AppendRow(r)
	}
	cmd.Println(t.Render())

	return nil
}
