package base

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"rig-cli",
	clientModule,
	fx.Provide(func() (*cmd_config.Config, error) {
		return cmd_config.NewConfig("")
	}),
	fx.Provide(zap.NewDevelopment),
	fx.Provide(getContext),
	fx.Provide(func(c *cmd_config.Context) *cmd_config.Auth {
		return c.GetAuth()
	}),
	fx.Provide(func(c *cmd_config.Context) *cmd_config.Service {
		return c.GetService()
	}),
	fx.Provide(func() context.Context { return context.Background() }),
	fx.Provide(func() (*client.Client, error) {
		return client.NewClientWithOpts(
			client.WithHostFromEnv(),
			client.WithAPIVersionNegotiation(),
		)
	}),
)

func getContext(cfg *cmd_config.Config) (*cmd_config.Context, error) {
	if cfg.CurrentContextName == "" {
		if len(cfg.Contexts) > 0 {
			fmt.Println("No context selected, please select one")
			if err := cmd_config.SelectContext(cfg); err != nil {
				return nil, err
			}
		} else {
			fmt.Println("No context available, please create one")
			if err := cmd_config.CreateDefaultContext(cfg); err != nil {
				return nil, err
			}
		}
	}

	c := cfg.GetCurrentContext()
	if c == nil {
		return nil, fmt.Errorf("no current context in config, run `rig config init`")
	}

	c.SetService(cfg.GetCurrentService())
	if c.GetService() == nil {
		return nil, fmt.Errorf("missing service config for context `%v`", cfg.CurrentContextName)
	}

	c.SetAuth(cfg.GetCurrentAuth())
	if c.GetAuth() == nil {
		return nil, fmt.Errorf("missing auth config for context `%v`", cfg.CurrentContextName)
	}

	return c, nil
}
