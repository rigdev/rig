package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/term"
)

// The way we use FX has some issues working with Cobra.
// Cobra works in the following stages:
//
// 1. You construct the entire Command structure and *the functions cobra needs to execute*
// 2. Cobra parses the command line arguments and find the command chain to execute
// 3. Cobra executes all PreRun functions
// 4. Cobra executes the main function
//
// We use FX to construct all dependencies needed for all the PreRuns + main function and group them
// into `Cmd` objects to encapsulate dependencies as member variables.
// For some dependencies, we may prompt the user during construction. This means we only want to build
// the exact set of dependencies needed for the PreRuns+main function to work.
// Unfortunately, Cobra:
// - Expects all functions it could potentially execute to exist before any parsing happens
// - Does not supply a step between parsing and execution of PreRun functions.
// We only have enough information to run FX after parsing and need to run FX before PreRuns are executed.
// This file contains functions which circumvent this issue.

var Module = fx.Module(
	"rig-cli",
	clientModule,
	fx.Provide(scheme.New),
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

func getContext(
	cfg *cmdconfig.Config,
	promptInfo *PromptInformation,
	interactive Interactive,
) (*cmdconfig.Context, error) {
	if cfg.CurrentContextName == "" {
		if interactive {
			if len(cfg.Contexts) > 0 {
				fmt.Println("No context selected, please select one")
				if err := cfg.SelectContext(); err != nil {
					return nil, err
				}
			} else {
				promptInfo.ContextCreation = true
				fmt.Println("No context available, please create one")
				if err := cfg.CreateDefaultContext(bool(interactive)); err != nil {
					return nil, err
				}
			}
		}
	}

	c := cfg.GetCurrentContext()
	if c == nil && !interactive {
		// No context configured. See if there is both host and auth available.
		if flags.Flags.Host == "" {
			return nil, fmt.Errorf("no host configured, use `--host` or `RIG_HOST` to specify the host of the Rig platform`")
		}

		if _, ok := os.LookupEnv("RIG_CLIENT_ID"); !ok {
			return nil, fmt.Errorf("missing RIG_CLIENT_ID environment variable")
		}

		if _, ok := os.LookupEnv("RIG_CLIENT_SECRET"); !ok {
			return nil, fmt.Errorf("missing RIG_CLIENT_SECRET environment variable")
		}

		flags.Flags.BasicAuth = true

		err := cfg.CreateContextNoPrompt("service-account", flags.Flags.Host)
		if err != nil {
			return nil, err
		}

		c = &cmdconfig.Context{}
		c.SetService(&cmdconfig.Service{
			Server: flags.Flags.Host,
		})
		c.SetAuth(&cmdconfig.Auth{})
	}
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

func AddOptions(opts ...fx.Option) {
	options = append(options, opts...)
}

func computeNumOfPreRuns(cmd *cobra.Command) int {
	res := 0
	for p := cmd; p != nil; p = p.Parent() {
		if p.PersistentPreRunE != nil {
			res++
		}
	}
	return res
}

func createOptions(cmd *cobra.Command, args []string) []fx.Option {
	return []fx.Option{
		Module,
		fx.NopLogger,
		fx.Provide(func() *cobra.Command { return cmd }),
		fx.Provide(func() []string { return args }),
		// provide a flag to indicate that we cannot prompt for resource creation
		fx.Provide(func() Interactive { return Interactive(term.IsTerminal(int(os.Stdin.Fd()))) }),
	}
}

func ExecuteInvokes(cmd *cobra.Command, args []string, invokes ...any) error {
	for _, invoke := range invokes {
		options = append(options, fx.Invoke(invoke))
	}
	allOpts := createOptions(cmd, args)
	allOpts = append(allOpts, options...)
	return fx.New(allOpts...).Err()
}

// PersistentPreRunE solves the issue described at the top of the file.
// It assumes all PreRunEs on the command chain has been wrapped in FX invokes
// and registered using InvokePreRunE.
// When Cobra starts executing PreRunEs, the following happens:
// If the current PreRunE being executed is not the last one, we simply do nothing.
// If the current PreRunE is the last one in the chain, we call FX to build all dependencies
// and run all Invokes.
// It is only at this point we know exactly which dependencies are needed. It is assumed the Cobra
// main Run function has had its dependencies initialized by one of the Invokes registered.
func PersistentPreRunE(cmd *cobra.Command, args []string) error {
	if firstPreRun {
		firstPreRun = false
		preRunsLeft = computeNumOfPreRuns(cmd)
	}
	preRunsLeft--

	if preRunsLeft > 0 || SkipFX(cmd) {
		return nil
	}

	allOpts := createOptions(cmd, args)
	allOpts = append(allOpts, options...)
	return fx.New(allOpts...).Err()
}

// IvokePreRunE registers FX invokes to be executed at the time a corresponding
// PreRunE would have been executed by Cobra had we not used FX.
func InvokePreRunE(cmd *cobra.Command, args []string, invokes ...any) error {
	for _, invoke := range invokes {
		options = append(options, fx.Invoke(invoke))
	}

	if err := PersistentPreRunE(cmd, args); err != nil {
		return err
	}
	return nil
}

// MakeInvokePreRunE constructs a PreRunE function signature which registers the
// supplied invokes to be executed at the time Cobra would have executed the returned PreRunE
// if we did not use FX at all.
func MakeInvokePreRunE(invokes ...any) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return InvokePreRunE(cmd, args, invokes...)
	}
}

type (
	FCtx = func(ctx context.Context, cmd *cobra.Command, args []string) error
	F    = func(cmd *cobra.Command, args []string) error
)

func CtxWrap(f FCtx) F {
	return func(cmd *cobra.Command, args []string) error {
		return f(context.Background(), cmd, args)
	}
}

type (
	FCompleteCtx = func(context.Context, *cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)
	FComplete    = func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)
)

func CtxWrapCompletion(f FCompleteCtx) FComplete {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return f(context.Background(), cmd, args, toComplete)
	}
}
