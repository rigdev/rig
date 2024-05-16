package version

import (
	"context"

	"connectrpc.com/connect"
	environment_api "github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/build"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Cmd struct {
	fx.In

	Rig    rig.Client
	Scope  scope.Scope
	Logger *zap.Logger
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
}

func Setup(parent *cobra.Command, s *cli.SetupContext, groupID string) {
	version := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(c *cobra.Command, args []string) error {
			if ok, _ := c.Flags().GetBool("full"); ok {
				if err := s.MakeInvokePreRunE(initCmd)(c, args); err != nil {
					return err
				}
			}
			// Honestly, a hack
			return cmd.version(context.Background(), c, args)
		},
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
		},
		GroupID: groupID,
	}
	version.Flags().BoolP("full", "v", false, "Print full version")
	parent.AddCommand(version)
}

func (c *Cmd) version(ctx context.Context, cmd *cobra.Command, _ []string) error {
	full, err := cmd.Flags().GetBool("full")
	if err != nil {
		return err
	}

	if full {
		cmd.Println(build.VersionStringFull())
	} else {
		cmd.Println(build.VersionString())
	}

	if full {
		resp, err := c.Rig.Environment().List(ctx, &connect.Request[environment_api.ListRequest]{})
		if err != nil {
			cmd.Println("Unable to get platform version", err)
		} else {
			cmd.Println("Platform version:", resp.Msg.GetPlatformVersion())
		}
	}

	return nil
}
