package cli

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	mockrig "github.com/rigdev/rig/gen/mocks/github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/spf13/afero"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func MakeTestModule(t *testing.T, isInteractive bool) fx.Option {
	return fx.Module(
		"test-rig-cli",
		fx.Provide(func() *testing.T {
			return t
		}),
		fx.Provide(func(t *testing.T) rig.Client {
			return mockrig.NewMockClient(t)
		}),
		fx.Provide(scheme.New),
		fx.Provide(func(fs afero.Fs) (*cmdconfig.Config, error) {
			return cmdconfig.NewConfig("", fs)
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
			return scope.Interactive(isInteractive)
		}),
	)
}
