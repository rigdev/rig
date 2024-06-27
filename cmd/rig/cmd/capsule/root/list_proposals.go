package root

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) listProposals(ctx context.Context, _ *cobra.Command, _ []string) error {
	capsuleID := capsule_cmd.CapsuleID
	resp, err := c.Rig.Capsule().ListProposals(ctx, connect.NewRequest(&capsule.ListProposalsRequest{
		ProjectId:     flags.GetProject(c.Scope),
		EnvironmentId: flags.GetEnvironment(c.Scope),
		CapsuleId:     capsuleID,
	}))
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetProposals(), flags.Flags.OutputType)
	}

	tbl := table.New("URL", "Age", "Creator").
		WithHeaderFormatter(color.New(color.FgBlue, color.Underline).SprintfFunc())
	for _, proposal := range resp.Msg.GetProposals() {
		age := common.FormatDuration(time.Since(proposal.GetMetadata().GetCreatedAt().AsTime()))
		tbl.AddRow(proposal.GetMetadata().GetReviewUrl(), age, proposal.GetMetadata().GetCreatedBy().GetPrintableName())
	}
	tbl.Print()

	return nil
}
