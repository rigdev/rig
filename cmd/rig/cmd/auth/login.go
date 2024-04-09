package auth

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func (c *Cmd) login(ctx context.Context, cmd *cobra.Command, _ []string) error {
	res, err := c.loginWithRetry(ctx, authUserIdentifier, authPassword)
	if err != nil {
		return err
	}

	uid, err := uuid.Parse(res.Msg.GetUserId())
	if err != nil {
		return err
	}

	c.Scope.GetCurrentContext().GetAuth().UserID = uid.String()
	c.Scope.GetCurrentContext().GetAuth().AccessToken = res.Msg.GetToken().GetAccessToken()
	c.Scope.GetCurrentContext().GetAuth().RefreshToken = res.Msg.GetToken().GetRefreshToken()
	if err := c.Scope.GetCfg().Save(); err != nil {
		return err
	}

	cmd.Println("Login successful!")

	return nil
}

func (c *Cmd) loginWithRetry(
	ctx context.Context,
	identifierStr, password string,
) (*connect.Response[authentication.LoginResponse], error) {
	shouldPromptIdentifier := identifierStr == ""
	shouldPromptPassword := password == ""
	var identifier *model.UserIdentifier
	for {
		var err error
		if shouldPromptIdentifier {
			identifier, err = c.Prompter.UserIndentifier()
		} else if identifier == nil {
			identifier, err = common.ParseUserIdentifier(authUserIdentifier)
		}
		if err != nil {
			return nil, err
		}

		if shouldPromptPassword {
			password, err = c.Prompter.Password("Enter Password:")
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
		}
		fmt.Println(err.Error())
	}
}
