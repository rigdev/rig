package noop

import (
	"fmt"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Prompter common.Prompter
}

var cc Cmd

func initCmd(c Cmd) {
	cc = c
}

type CmdNoRig struct {
	fx.In

	Prompter common.Prompter
}

var ccNoRig CmdNoRig

func initCmdNoRig(c CmdNoRig) {
	ccNoRig = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	cmd := &cobra.Command{
		Use: "noop",
	}
	cmd1 := &cobra.Command{
		Use:               "cmd1",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitUser:        "",
			auth.OmitEnvironment: "",
			auth.OmitCapsule:     "",
			auth.OmitProject:     "",
		},
		RunE: cc.noop,
	}
	cmd.AddCommand(cmd1)

	cmd2 := &cobra.Command{
		Use:               "cmd2",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		RunE:              cc.noop,
	}
	cmd.AddCommand(cmd2)

	cmd3 := &cobra.Command{
		Use:               "cmd3",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdNoRig),
		Annotations: map[string]string{
			auth.OmitUser:        "",
			auth.OmitEnvironment: "",
			auth.OmitCapsule:     "",
			auth.OmitProject:     "",
		},
		RunE: ccNoRig.noop,
	}
	cmd.AddCommand(cmd3)

	parent.AddCommand(cmd)
}

func (c *Cmd) noop(_ *cobra.Command, _ []string) error {
	fmt.Println("noop")
	return nil
}

func (c *CmdNoRig) noop(_ *cobra.Command, _ []string) error {
	fmt.Println("noop")
	return nil
}
