package auth

import (
	"context"
	"os"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/spf13/cobra"
)

func initServiceAccountContext(cfg *cmdconfig.Config) (cli.ContextDependency, error) {
	contextName := flags.Flags.Context
	if contextName == "" {
		contextName = "service-account"
	}

	rCtx := cfg.GetContext(contextName)
	if rCtx == nil {
		if err := cfg.CreateContext(contextName, flags.Flags.Host, false); err != nil {
			return nil, err
		}

		cfg.CurrentContextName = contextName
	}

	return struct{}{}, nil
}

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

	rCtx := c.Scope.GetCurrentContext()
	rCtx.GetAuth().UserID = os.Getenv("RIG_CLIENT_ID")
	rCtx.GetAuth().AccessToken = res.Msg.GetToken().GetAccessToken()
	rCtx.GetAuth().RefreshToken = res.Msg.GetToken().GetRefreshToken()
	if err := c.Scope.GetCfg().Save(); err != nil {
		return err
	}

	cmd.Println("Service Account activated!")

	return nil
}
