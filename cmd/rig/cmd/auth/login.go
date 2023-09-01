package auth

import (
	"context"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func AuthLogin(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client, cfg *base.Config, logger *zap.Logger) error {
	if authUser == "" {
		u, err := utils.PromptGetInput("Enter Username or Email", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}

		authUser = u
	}

	var id *model.UserIdentifier
	if strings.Contains(authUser, "@") {
		id = &model.UserIdentifier{
			Identifier: &model.UserIdentifier_Email{
				Email: authUser,
			},
		}
	} else {
		id = &model.UserIdentifier{
			Identifier: &model.UserIdentifier_Username{
				Username: authUser,
			},
		}
	}

	if authPassword == "" {
		pw, err := utils.GetPasswordPrompt("Enter Password")
		if err != nil {
			return err
		}

		authPassword = string(pw)
	}

	res, err := nc.Authentication().Login(ctx, &connect.Request[authentication.LoginRequest]{
		Msg: &authentication.LoginRequest{
			Method: &authentication.LoginRequest_UserPassword{
				UserPassword: &authentication.UserPassword{
					Identifier: id,
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
