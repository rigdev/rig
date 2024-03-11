package auth

import (
	"context"
	"os"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/spf13/cobra"
)

func (c *Cmd) activateServiceAccount(ctx context.Context, cmd *cobra.Command, _ []string) error {
	res, err := c.Rig.Authentication().Login(ctx, &connect.Request[authentication.LoginRequest]{
		Msg: &authentication.LoginRequest{
			Method: &authentication.LoginRequest_ClientCredentials{
				ClientCredentials: &authentication.ClientCredentials{
					ClientId:     os.Getenv("RIG_CLIENT_ID"),
					ClientSecret: os.Getenv("RIG_CLIENT_SECRET"),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	c.Cfg.GetCurrentAuth().UserID = os.Getenv("RIG_CLIENT_ID")
	c.Cfg.GetCurrentAuth().AccessToken = res.Msg.GetToken().GetAccessToken()
	c.Cfg.GetCurrentAuth().RefreshToken = res.Msg.GetToken().GetRefreshToken()
	if err := c.Cfg.Save(); err != nil {
		return err
	}

	cmd.Println("Service Account activated!")

	return nil
}
