package auth

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func AuthLogin(ctx context.Context, cmd *cobra.Command, client rig.Client, cfg *base.Config) error {
	var identifier *model.UserIdentifier
	var err error
	if authUserIdentifier == "" {
		identifier, err = common.PromptUserIndentifier()
		if err != nil {
			return err
		}
	} else {
		identifier, err = common.ParseUserIdentifier(authUserIdentifier)
	}

	if authPassword == "" {
		pw, err := common.GetPasswordPrompt("Enter Password")
		if err != nil {
			return err
		}
		authPassword = string(pw)
	}

	res, err := client.Authentication().Login(ctx, &connect.Request[authentication.LoginRequest]{
		Msg: &authentication.LoginRequest{
			Method: &authentication.LoginRequest_UserPassword{
				UserPassword: &authentication.UserPassword{
					Identifier: identifier,
					Password:   authPassword,
					ProjectId:  auth.RigProjectID.String(),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	uid, err := uuid.Parse(res.Msg.GetUserId())
	if err != nil {
		return err
	}

	cfg.Auth().UserID = uid
	cfg.Auth().AccessToken = res.Msg.GetToken().GetAccessToken()
	cfg.Auth().RefreshToken = res.Msg.GetToken().GetRefreshToken()
	if err := cfg.Save(); err != nil {
		return err
	}

	cmd.Println("Login successful!")

	return nil
}
