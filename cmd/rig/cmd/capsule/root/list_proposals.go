package root

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
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
	return printProposals(resp.Msg.GetProposals())
}

func (c *Cmd) listSetProposals(ctx context.Context, _ *cobra.Command, _ []string) error {
	capsuleID := capsule_cmd.CapsuleID
	resp, err := c.Rig.Capsule().ListSetProposals(ctx, connect.NewRequest(&capsule.ListSetProposalsRequest{
		ProjectId: flags.GetProject(c.Scope),
		CapsuleId: capsuleID,
	}))
	if err != nil {
		return err
	}
	return printProposals(resp.Msg.GetProposals())
}

func printProposals[T interface {
	GetMetadata() *model.ProposalMetadata
}](proposals []T) error {
	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(proposals, flags.Flags.OutputType)
	}

	tbl := table.New("URL", "Age", "Creator").
		WithHeaderFormatter(color.New(color.FgBlue, color.Underline).SprintfFunc())
	for _, proposal := range proposals {
		age := common.FormatDuration(time.Since(proposal.GetMetadata().GetCreatedAt().AsTime()))
		tbl.AddRow(proposal.GetMetadata().GetReviewUrl(), age, proposal.GetMetadata().GetCreatedBy().GetPrintableName())
	}
	tbl.Print()

	return nil
}
