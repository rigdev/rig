package main

import (
	"github.com/rigdev/rig/cmd/rig/cmd"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		base.Module,
		fx.NopLogger,
		fx.Invoke(func(r cmd.RootCmd) error {
			return r.Execute()
		}),
	)
}
