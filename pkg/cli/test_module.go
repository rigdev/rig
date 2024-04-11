package cli

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/spf13/afero"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type TestModuleInput struct {
	RigClient     rig.Client
	IsInteractive bool
	Prompter      common.Prompter
	FS            afero.Fs
}

func MakeTestModule(i TestModuleInput) fx.Option {
	return fx.Module(
		"test-rig-cli",
		fx.Provide(func() afero.Fs { return i.FS }),
		fx.Provide(func() common.Prompter { return i.Prompter }),
		fx.Provide(func(cfg *cmdconfig.Config) (rig.Client, error) {
			_, err := getClientOptions(cfg)
			if err != nil {
				return nil, err
			}

			return i.RigClient, nil
		}),
		fx.Provide(auth.NewService),
		fx.Provide(scheme.New),
		fx.Provide(func(fs afero.Fs, p common.Prompter) (*cmdconfig.Config, error) {
			return cmdconfig.NewConfig("", fs, p)
		}),
		fx.Provide(zap.NewDevelopment),
		fx.Provide(getContext),
		fx.Provide(scope.NewScope),
		fx.Provide(func() context.Context { return context.Background() }),
		fx.Provide(func() (*client.Client, error) {
			return nil, nil // TODO
		}),
		fx.Provide(func() *PromptInformation { return &PromptInformation{} }),
		fx.Provide(func() scope.Interactive {
			return scope.Interactive(i.IsInteractive)
		}),
	)
}
