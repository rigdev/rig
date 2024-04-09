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
		fx.Provide(func() rig.Client { return i.RigClient }),
		fx.Provide(auth.NewService),
		fx.Invoke(authRigClient),
		fx.Provide(scheme.New),
		fx.Provide(func(fs afero.Fs, p common.Prompter) (*cmdconfig.Config, error) {
			return cmdconfig.NewConfig("", fs, p)
		}),
		fx.Provide(zap.NewDevelopment),
		fx.Provide(getContext),
		fx.Provide(scope.NewScope),
		fx.Supply(context.Background()),
		fx.Provide(func() (*client.Client, error) {
			return nil, nil // TODO
		}),
		fx.Provide(func() *PromptInformation { return &PromptInformation{} }),
		fx.Provide(func() scope.Interactive {
			return scope.Interactive(i.IsInteractive)
		}),
	)
}
