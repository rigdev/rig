package base

import (
	"context"
	"fmt"

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
		cfg, err := NewConfig("")
		if err != nil {
			return err
		}

		f := fx.New(
			clientModule,
			fx.Supply(cfg),
			fx.Supply(cmd),
			fx.Supply(args),
			fx.Provide(zap.NewDevelopment),
			fx.Provide(func() (*Context, error) {
				if cfg.CurrentContext == "" {
					if len(cfg.Contexts) > 0 {
						cmd.Println("No context selected, please select one")
						if err := SelectContext(cfg); err != nil {
							return nil, err
						}
					} else {
						cmd.Println("No context available, please create one")
						if err := CreateContext(cfg); err != nil {
							return nil, err
						}
					}
				}

				c := cfg.Context()
				if c == nil {
					return nil, fmt.Errorf("no current context in config, run `rig config init`")
				}

				c.service = cfg.Service()
				if c.service == nil {
					return nil, fmt.Errorf("missing service config for context `%v`", cfg.CurrentContext)
				}

				c.auth = cfg.Auth()
				if c.auth == nil {
					return nil, fmt.Errorf("missing auth config for context `%v`", cfg.CurrentContext)
				}

				return c, nil
			}),
			fx.Provide(func(c *Context) *Auth {
				return c.auth
			}),
			fx.Provide(func(c *Context) *Service {
				return c.service
			}),
			fx.Provide(func() context.Context { return context.Background() }),
			fx.Options(_options...),
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
