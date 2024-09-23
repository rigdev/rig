package root

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, _ []string) error {
	resp, err := c.Rig.Capsule().Get(ctx, &connect.Request[capsule.GetRequest]{
		Msg: &capsule.GetRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			ProjectId: c.Scope.GetCurrentContext().GetProject(),
		},
	})
	if err != nil {
		return err
	}

	var environments []*platformv1.Capsule
	for _, env := range resp.Msg.GetEnvironmentRevisions() {
		environments = append(environments, env.GetSpec())
	}
	slices.SortFunc(environments, func(e1, e2 *platformv1.Capsule) int {
		return strings.Compare(e1.GetEnvironment(), e2.GetEnvironment())
	})

	var out any
	if flags.Flags.Environment != "" {
		for _, env := range environments {
			if env.GetEnvironment() == flags.Flags.Environment {
				out = env
				break
			}
		}
	} else {
		out = environments
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		if out == nil {
			fmt.Println("capsule not deployed in environment", flags.Flags.Environment)
			return nil
		}
		return common.FormatPrint(out, flags.Flags.OutputType)
	}

	cc := resp.Msg.GetCapsule()

	type output struct {
		Capsule *capsule.Capsule `json:"capsule" yaml:"capsule"`
		Rollout *capsule.Rollout `json:"rollout" yaml:"rollout"`
	}

	var op output

	r, err := capsule_cmd.GetCurrentRolloutOfCapsule(ctx, c.Rig, c.Scope, cc.GetCapsuleId())
	if errors.IsNotFound(err) {
		// OK, default values.
		r = &capsule.Rollout{}
	} else if err != nil {
		return err
	}

	op = output{
		Capsule: cc,
		Rollout: r,
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"ID", "Replicas", "Image ID"})
	t.AppendRow(table.Row{
		op.Capsule.GetCapsuleId(),
		op.Rollout.GetConfig().GetReplicas(),
		op.Rollout.GetConfig().GetImageId(),
	})

	cmd.Println(t.Render())

	return nil
}
