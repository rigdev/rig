package auth

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func (c Cmd) login(ctx context.Context, cmd *cobra.Command, args []string) error {
	res, err := c.loginWithRetry(ctx, authUserIdentifier, authPassword, auth.RigProjectID)
	if err != nil {
		return err
	}

	uid, err := uuid.Parse(res.Msg.GetUserId())
	if err != nil {
		return err
	}

	c.Cfg.GetCurrentAuth().UserID = uid
	c.Cfg.GetCurrentAuth().AccessToken = res.Msg.GetToken().GetAccessToken()
	c.Cfg.GetCurrentAuth().RefreshToken = res.Msg.GetToken().GetRefreshToken()
	if err := c.Cfg.Save(); err != nil {
		return err
	}

	cmd.Println("Login successful!")

	return nil
}

func (c Cmd) loginWithRetry(ctx context.Context, identifierStr, password, project string) (*connect.Response[authentication.LoginResponse], error) {
	shouldPromptIdentifier := identifierStr == ""
	shouldPromptPassword := password == ""
	var identifier *model.UserIdentifier
	for {
		var err error
		if shouldPromptIdentifier {
			identifier, err = common.PromptUserIndentifier()
		} else if identifier == nil {
			identifier, err = common.ParseUserIdentifier(authUserIdentifier)
		}
		if err != nil {
			return nil, err
		}

		if shouldPromptPassword {
			password, err = common.PromptPassword("Enter Password:")
			if err != nil {
				return nil, err
			}
		}

		res, err := c.Rig.Authentication().Login(ctx, &connect.Request[authentication.LoginRequest]{
			Msg: &authentication.LoginRequest{
				Method: &authentication.LoginRequest_UserPassword{
					UserPassword: &authentication.UserPassword{
						Identifier: identifier,
						Password:   password,
						ProjectId:  project,
					},
				},
			},
		})
		if err == nil {
			return res, nil
		}

		if errors.IsNotFound(err) {
			if !shouldPromptIdentifier {
				return nil, err
			}
			fmt.Println("User not found")
			continue
		} else if errors.IsUnauthenticated(err) {
			if !shouldPromptPassword {
				return nil, err
			}
			shouldPromptIdentifier = false
			fmt.Println("Wrong password")
			continue
		} else {
			fmt.Println(err.Error())
		}
	}
}
