package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/spf13/afero"
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

type SetupContext struct {
	BaseModule     fx.Option
	Args           []string
	AddTestCommand bool
	options        []fx.Option
	firstPreRun    bool
	preRunsLeft    int
}

func NewSetupContext(baseModule fx.Option, args []string) *SetupContext {
	return &SetupContext{
		BaseModule:  baseModule,
		Args:        args,
		options:     nil,
		firstPreRun: true,
		preRunsLeft: 0,
	}
}

var StandardSetupContext = NewSetupContext(Module, nil)

var Module = fx.Module(
	"rig-cli",
	clientModule,
	fx.Provide(afero.NewOsFs),
	fx.Provide(scheme.New),
	fx.Provide(func(fs afero.Fs, prompter common.Prompter) (*cmdconfig.Config, error) {
		return cmdconfig.NewConfig("", fs, prompter)
	}),
	fx.Provide(zap.NewDevelopment),
	fx.Provide(getContext),
	fx.Provide(scope.NewScope),
	fx.Provide(func() context.Context { return context.Background() }),
	fx.Provide(func() (*client.Client, error) {
		return client.NewClientWithOpts(
			client.WithHostFromEnv(),
			client.WithAPIVersionNegotiation(),
		)
	}),
	fx.Provide(func() *PromptInformation { return &PromptInformation{} }),
	// provide a flag to indicate that we cannot prompt for resource creation
	fx.Provide(func() scope.Interactive { return scope.Interactive(term.IsTerminal(int(os.Stdin.Fd()))) }),
	fx.Provide(func() common.Prompter { return common.StandardPrompter{} }),
)

func getContext(
	cfg *cmdconfig.Config,
	promptInfo *PromptInformation,
	interactive scope.Interactive,
) (*cmdconfig.Context, error) {
	if cfg.CurrentContextName == "" && flags.Flags.Context == "" {
		if interactive {
			if len(cfg.Contexts) > 0 {
				fmt.Println("No context selected, please select one")
				if err := cfg.SelectContext(); err != nil {
					return nil, err
				}
			} else {
				promptInfo.ContextCreation = true
				fmt.Println("No context available, please create one")
				if err := cfg.CreateContext("", "", bool(interactive)); err != nil {
					return nil, err
				}
			}
		}
	}
	c := cfg.GetCurrentContext()
	if flags.Flags.Context != "" {
		found := false
		for _, context := range cfg.Contexts {
			if context.Name == flags.Flags.Context {
				found = true
				c = context
			}
		}
		if !found {
			return nil, fmt.Errorf("context `%v` not found", flags.Flags.Context)
		}
	}
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

		c = cfg.GetCurrentContext()
	}
	if c == nil {
		// This shouldn't happen as we prompt for a config if one is missing above
		return nil, fmt.Errorf("no current context in config, run `rig config init`")
	}

	service, err := cfg.GetService(c.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("missing service config for context `%v`", cfg.CurrentContextName)
	}
	c.SetService(service)

	user, err := cfg.GetUser(c.Name)
	if err != nil {
		return nil, fmt.Errorf("missing user config for context `%v`", cfg.CurrentContextName)
	}
	c.SetAuth(user.Auth)

	return c, nil
}

func (s *SetupContext) AddOptions(opts ...fx.Option) {
	s.options = append(s.options, opts...)
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

func (s *SetupContext) createOptions(cmd *cobra.Command, args []string) []fx.Option {
	return []fx.Option{
		s.BaseModule,
		fx.NopLogger,
		fx.Provide(func() *cobra.Command { return cmd }),
		fx.Provide(func() []string { return args }),
	}
}

func (s *SetupContext) ExecuteInvokes(cmd *cobra.Command, args []string, invokes ...any) error {
	for _, invoke := range invokes {
		s.options = append(s.options, fx.Invoke(invoke))
	}
	allOpts := s.createOptions(cmd, args)
	allOpts = append(allOpts, s.options...)
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
func (s *SetupContext) PersistentPreRunE(cmd *cobra.Command, args []string) error {
	if s.firstPreRun {
		s.firstPreRun = false
		s.preRunsLeft = computeNumOfPreRuns(cmd)
	}
	s.preRunsLeft--

	if s.preRunsLeft > 0 || SkipFX(cmd) {
		return nil
	}

	allOpts := s.createOptions(cmd, args)
	allOpts = append(allOpts, s.options...)
	return fx.New(allOpts...).Err()
}

// IvokePreRunE registers FX invokes to be executed at the time a corresponding
// PreRunE would have been executed by Cobra had we not used FX.
func (s *SetupContext) InvokePreRunE(cmd *cobra.Command, args []string, invokes ...any) error {
	for _, invoke := range invokes {
		s.options = append(s.options, fx.Invoke(invoke))
	}

	if err := s.PersistentPreRunE(cmd, args); err != nil {
		return err
	}
	return nil
}

// MakeInvokePreRunE constructs a PreRunE function signature which registers the
// supplied invokes to be executed at the time Cobra would have executed the returned PreRunE
// if we did not use FX at all.
func (s *SetupContext) MakeInvokePreRunE(invokes ...any) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return s.InvokePreRunE(cmd, args, invokes...)
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
	//nolint:lll
	FCompleteCtxSetup = func(context.Context, *cobra.Command, []string, string, *SetupContext) ([]string, cobra.ShellCompDirective)
	FCompleteSetup    = func(*cobra.Command, []string, string, *SetupContext) ([]string, cobra.ShellCompDirective)
)

func CtxWrapCompletion(f FCompleteCtx) FComplete {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return f(context.Background(), cmd, args, toComplete)
	}
}

func HackCtxWrapCompletion(f FCompleteCtxSetup, s *SetupContext) FComplete {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return f(context.Background(), cmd, args, toComplete, s)
	}
}

func HackWrapCompletion(f FCompleteSetup, s *SetupContext) FComplete {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return f(cmd, args, toComplete, s)
	}
}
