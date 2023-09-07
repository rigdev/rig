package fxtest

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	someFlag string
)

type subcmd struct {
	fx.In

	ctx    context.Context
	cfg    *base.Config
	client rig.Client
}

func (f *subcmd) Setup(parent *cobra.Command) {
	fxtest := &cobra.Command{
		Use:   "subcmd",
		Short: "subcmd test some stuff",
		Args:  cobra.ExactArgs(1),
		RunE:  f.Run,
	}
	fxtest.Flags().StringVarP(&someFlag, "someflag", "s", "", "some flag")

	parent.AddCommand(fxtest)
}

func (f *subcmd) Run(cmd *cobra.Command, args []string) error {
	return nil
}

type root struct {
	subcmd subcmd
}

func (r *root) Execute() {
	cmd := &cobra.Command{
		Use:   "fxtest",
		Short: "test of fx + cobra",
	}

	r.subcmd.Setup(cmd)
	return cmd.Execute()
}
