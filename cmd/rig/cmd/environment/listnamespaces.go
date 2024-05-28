package environment

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) listNamespaces(ctx context.Context, _ *cobra.Command, args []string) error {
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		if !c.Scope.IsInteractive() {
			return fmt.Errorf("missing environment argument")
		}
		resp, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
		if err != nil {
			return err
		}
		var names []string
		for _, e := range resp.Msg.GetEnvironments() {
			names = append(names, e.GetEnvironmentId())
		}
		if _, name, err = c.Prompter.Select("Environment", names); err != nil {
			return err
		}
	}

	resp, err := c.Rig.Environment().GetNamespaces(ctx, connect.NewRequest(&environment.GetNamespacesRequest{}))
	if err != nil {
		return err
	}

	var res []*environment.ProjectEnvironmentNamespace
	for _, ns := range resp.Msg.GetNamespaces() {
		if ns.GetEnvironmentId() == name {
			res = append(res, ns)
		}
	}

	if base.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(res, base.Flags.OutputType)
	}

	tbl := table.
		New("Project", "Environment", "Namespaces").
		WithHeaderFormatter(color.New(color.FgBlue, color.Underline).SprintfFunc())
	for _, ns := range res {
		tbl.AddRow(ns.GetProjectId(), ns.GetEnvironmentId(), ns.GetNamespace())
	}
	tbl.Print()

	return nil
}
