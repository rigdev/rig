package base

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var _options []fx.Option

func AddOptions(opts ...fx.Option) {
	_options = append(_options, opts...)
}

func Register(f interface{}) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cfg, err := cmd_config.NewConfig("")
		if err != nil {
			return err
		}

		f := fx.New(
			clientModule,
			fx.Supply(cfg),
			fx.Supply(cmd),
			fx.Supply(args),
			fx.Provide(zap.NewDevelopment),
			fx.Provide(getContext),
			fx.Provide(func(c *cmd_config.Context) *cmd_config.Auth {
				return c.GetAuth()
			}),
			fx.Provide(func(c *cmd_config.Context) *cmd_config.Service {
				return c.GetService()
			}),
			fx.Provide(func() context.Context { return context.Background() }),
			fx.Options(_options...),
			fx.Provide(func() (*client.Client, error) {
				return client.NewClientWithOpts(
					client.WithHostFromEnv(),
					client.WithAPIVersionNegotiation(),
				)
			}),
			fx.Invoke(CheckAuth),
			fx.Invoke(f),
			fx.NopLogger,
		)

		if err := f.Start(context.Background()); err != nil {
			return dig.RootCause(err)
		}
		if err := f.Stop(context.Background()); err != nil {
			return dig.RootCause(err)
		}
		return dig.RootCause(f.Err())
	}
}

func getContext(cfg *cmd_config.Config, cmd *cobra.Command) (*cmd_config.Context, error) {
	if cfg.CurrentContextName == "" {
		if len(cfg.Contexts) > 0 {
			cmd.Println("No context selected, please select one")
			if err := cmd_config.SelectContext(cfg); err != nil {
				return nil, err
			}
		} else {
			cmd.Println("No context available, please create one")
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
