package base

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
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

// Register solves an annoying problem I haven't found a good solution to yet.
// 1. We want to use FX to generate all the dependencies for the commands
//
// 2. We only want to generate dependencies strictly for the commands currently in the
// command chain being executed. Some dependencies prompt the user interactively if their
// values don't already exist (e.g. some Config and Context does this). We don't want this
// to happen for all commands  (otherwise it would prompt during e.g. 'help' which is super weird behaviour).
//
// 3. We would like to encapsulate all dependencies to a command in a struct (called Cmd structs here)
// to make it possible to easily call helper functions with the dependencies as well and not have
// unwieldy function signatures.
//
// 4. Cobra has no lifecycle step in between parsing all commands and starting executing the functions
// defined on the cobra.Commands (e.g. PreRun, Run...)
//
// Conclusion: We need to have a reference to all functions being called before any cobra parsing happens
// These functions are instance methods on objects we cannot construct before any Cobra parsing happens
// :(
func Register[T any](f func(T) any) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fx.New(
			Module,
			fx.NopLogger,
			fx.Provide(func() *cobra.Command { return cmd }),
			fx.Provide(func() []string { return args }),
			fx.Invoke(func(t T, ctx context.Context, cmd *cobra.Command, args []string) error {
				switch f := f(t).(type) {
				case func(context.Context, *cobra.Command, []string) error:
					return f(ctx, cmd, args)
				case func(*cobra.Command, []string) error:
					return f(cmd, args)
				default:
					return fmt.Errorf("unexpected function signature %T to Register", f)
				}
			}),
		).Err()
	}
}

// See Register
func RegisterCompletion[T any](f func(T) any) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		var directive cobra.ShellCompDirective
		if err := fx.New(
			Module,
			fx.NopLogger,
			fx.Provide(func() *cobra.Command { return cmd }),
			fx.Provide(func() []string { return args }),
			fx.Invoke(func(t T, ctx context.Context, cmd *cobra.Command, args []string) error {
				switch f := f(t).(type) {
				case func(*cobra.Command, []string, string) error:
					return f(cmd, args, toComplete)
				case func(context.Context, *cobra.Command, []string, string) error:
					return f(ctx, cmd, args, toComplete)
				default:
					return fmt.Errorf("unexpected function signature %T to Register", f)
				}
			}),
		).Err(); err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return completions, directive
	}
}
