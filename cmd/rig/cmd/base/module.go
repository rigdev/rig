package base

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"rig-cli",
	clientModule,
	fx.Provide(func() (*cmdconfig.Config, error) {
		return cmdconfig.NewConfig("")
	}),
	fx.Provide(zap.NewDevelopment),
	fx.Provide(getContext),
	fx.Provide(func(c *cmdconfig.Context) *cmdconfig.Auth {
		return c.GetAuth()
	}),
	fx.Provide(func(c *cmdconfig.Context) *cmdconfig.Service {
		return c.GetService()
	}),
	fx.Provide(func() context.Context { return context.Background() }),
	fx.Provide(func() (*client.Client, error) {
		return client.NewClientWithOpts(
			client.WithHostFromEnv(),
			client.WithAPIVersionNegotiation(),
		)
	}),
	fx.Provide(func() *PromptInformation { return &PromptInformation{} }),
)

type Interactive bool

func skipContext(cmd *cobra.Command) bool {
	annotations := common.GetAllAnnotations(cmd)
	if flags.Flags.Host == "" {
		return false
	}

	if _, ok := annotations[auth.OmitUser]; !ok && !flags.Flags.BasicAuth {
		return false
	}

	if _, ok := annotations[auth.OmitProject]; !ok && flags.Flags.Project == "" {
		return false
	}

	if _, ok := annotations[auth.OmitEnvironment]; !ok && flags.Flags.Environment == "" {
		return false
	}

	return true
}

func getContext(cmd *cobra.Command, cfg *cmdconfig.Config, promptInfo *PromptInformation) (*cmdconfig.Context, error) {
	if skipContext(cmd) {
		return &cmdconfig.Context{}, nil
	}

	if cfg.CurrentContextName == "" {
		if len(cfg.Contexts) > 0 {
			fmt.Println("No context selected, please select one")
			if err := cfg.SelectContext(); err != nil {
				return nil, err
			}
		} else {
			promptInfo.ContextCreation = true
			fmt.Println("No context available, please create one")
			if err := cfg.CreateDefaultContext(); err != nil {
				return nil, err
			}
		}
	}

	c := cfg.GetCurrentContext()
	if c == nil {
		// This shouldn't happen as we prompt for a config if one is missing above
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
			res++
		}
	}
	return res
}

func Provide(cmd *cobra.Command, args []string, invokes ...any) error {
	for _, invoke := range invokes {
		options = append(options, fx.Invoke(invoke))
	}

	allOpts := []fx.Option{
		Module,
		fx.NopLogger,
		fx.Provide(func() *cobra.Command { return cmd }),
		fx.Provide(func() []string { return args }),
		// provide a flag to indicate that we cannot prompt for resource creation
		fx.Provide(func() Interactive { return false }),
	}
	allOpts = append(allOpts, options...)
	return fx.New(allOpts...).Err()
}

func PersistentPreRunE(cmd *cobra.Command, args []string) error {
	if firstPreRun {
		firstPreRun = false
		preRunsLeft = computeNumOfPreRuns(cmd)
	}
	preRunsLeft--

	if preRunsLeft == 0 && !SkipChecks(cmd) {
		allOpts := []fx.Option{
			Module,
			fx.NopLogger,
			fx.Provide(func() *cobra.Command { return cmd }),
			fx.Provide(func() []string { return args }),
			// provide a flag to indicate that we can prompt for resource creation
			fx.Provide(func() Interactive { return true }),
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
