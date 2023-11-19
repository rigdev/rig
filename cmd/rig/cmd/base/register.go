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

var (
	options     []fx.Option
	firstPreRun = true
	preRunsLeft = 0
)

func computeNumOfPreRuns(cmd *cobra.Command) int {
	res := 0
	for p := cmd; p != nil; p = p.Parent() {
		if p.PersistentPreRunE != nil {
			res += 1
		}
	}
	return res
}

func PersistentPreRunE(cmd *cobra.Command, args []string) error {
	if firstPreRun {
		firstPreRun = false
		preRunsLeft = computeNumOfPreRuns(cmd)
	}
	preRunsLeft -= 1

	if preRunsLeft == 0 && !skipChecks(cmd) {
		allOpts := []fx.Option{
			Module,
			fx.NopLogger,
			fx.Provide(func() *cobra.Command { return cmd }),
			fx.Provide(func() []string { return args }),
		}
		allOpts = append(allOpts, options...)
		return fx.New(allOpts...).Err()
	}
	return nil
}

func InvokePreRunE(cmd *cobra.Command, args []string, invokes ...any) error {
	for _, invoke := range invokes {
		options = append(options, fx.Invoke(invoke))
	}

	if err := PersistentPreRunE(cmd, args); err != nil {
		return err
	}
	return nil
}

func MakeInvokePreRunE(fs ...any) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return InvokePreRunE(cmd, args, fs...)
	}
}

type FCtx = func(ctx context.Context, cmd *cobra.Command, args []string) error
type F = func(cmd *cobra.Command, args []string) error

func CtxWrap(f FCtx) F {
	return func(cmd *cobra.Command, args []string) error {
		return f(context.Background(), cmd, args)
	}
}

type FCompleteCtx = func(context.Context, *cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)
type FComplete = func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)

func CtxWrapCompletion(f FCompleteCtx) FComplete {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return f(context.Background(), cmd, args, toComplete)
	}
}
