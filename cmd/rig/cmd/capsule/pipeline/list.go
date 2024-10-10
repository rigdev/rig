package pipeline

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	pipeline_api "github.com/rigdev/rig-go-api/api/v1/capsule/pipeline"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	table2 "github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) list(ctx context.Context, _ *cobra.Command, _ []string) error {
	listResp, err := c.Rig.Capsule().ListPipelineStatuses(ctx, connect.NewRequest(&capsule.ListPipelineStatusesRequest{
		Pagination: &model.Pagination{
			Offset:     uint32(offset),
			Limit:      uint32(limit),
			Descending: true,
		},
		CapsuleFilter: capsule_cmd.CapsuleID,
		ProjectFilter: c.Scope.GetCurrentContext().GetProject(),
		NameFilter:    pipelineName,
	}))
	if err != nil {
		return err
	}

	statuses := listResp.Msg.GetStatuses()

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(statuses, flags.Flags.OutputType)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table2.New("ID", "Pipeline", "Started", "State", "Current Phase",
		"Phase State", "Phase Rollout", "Phase Message")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	var rows [][]string
	for _, s := range statuses {
		rows = append(rows, pipelineStatusToTableRow(s))
	}
	tbl.SetRows(rows)
	tbl.Print()

	return nil
}

func pipelineStatusToTableRow(s *pipeline_api.Status) []string {
	currentPhaseStatus := s.GetPhaseStatuses()[s.CurrentPhase]
	msg := ""
	if len(currentPhaseStatus.GetMessages()) > 0 {
		msg = currentPhaseStatus.GetMessages()[len(currentPhaseStatus.GetMessages())-1].Message
	}

	return []string{
		fmt.Sprint(s.GetExecutionId()),
		s.GetPipelineName(),
		common.FormatTime(s.GetStartedAt().AsTime()),
		s.GetState().String(),
		currentPhaseStatus.GetEnvironmentId(),
		currentPhaseStatus.GetState().String(),
		fmt.Sprint(currentPhaseStatus.GetRolloutId()),
		msg,
	}
}
