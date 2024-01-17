package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Cmd) executions(ctx context.Context, _ *cobra.Command, _ []string) error {
	from, to, err := parseFromTo()
	if err != nil {
		return err
	}

	states, err := parseStates()
	if err != nil {
		return err
	}

	resp, err := c.Rig.Capsule().GetJobExecutions(ctx, connect.NewRequest(&capsule.GetJobExecutionsRequest{
		CapsuleId:   capsule_cmd.CapsuleID,
		JobName:     jobName,
		States:      states,
		CreatedFrom: timeToPB(from),
		CreatedTo:   timeToPB(to),
		Pagination: &model.Pagination{
			Limit:      limit,
			Descending: true,
		},
		ProjectId:     c.Cfg.GetProject(),
		EnvironmentId: base.GetEnvironment(c.Cfg),
	}))
	if err != nil {
		return err
	}

	executions := resp.Msg.GetJobExecutions()
	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(executions)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("#", "Name", "Created", "Finished", "State", "Retries")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for idx, e := range executions {
		row := []any{
			idx + 1,
			e.GetJobName(),
			formatTime(e.GetCreatedAt().AsTime()),
			formatTime(e.GetFinishedAt().AsTime()),
			stateToStr(e.GetState()),
			e.GetRetries(),
		}
		tbl.AddRow(row...)
	}
	tbl.Print()

	return nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
}

func timeToPB(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func parseFromTo() (time.Time, time.Time, error) {
	var to, from time.Time

	if (toStr != "" || fromStr != "") && since != "" {
		return from, to, errors.New("either --from/--to or --since can be given, not both")
	}

	if err := parseTime(fromStr, &from); err != nil {
		return from, to, fmt.Errorf("--from malformed: %s", err)
	}
	if err := parseTime(toStr, &to); err != nil {
		return from, to, fmt.Errorf("--to malformed: %s", err)
	}
	if since != "" {
		sinceDuration, err := time.ParseDuration(since)
		if err != nil {
			return from, to, fmt.Errorf("--since malformed: %s", since)
		}
		from = time.Now().Add(-sinceDuration)
	}
	return from, to, nil
}

func stateToStr(s capsule.JobState) string {
	switch s {
	case capsule.JobState_JOB_STATE_UNSPECIFIED:
		return "unspecified"
	case capsule.JobState_JOB_STATE_ONGOING:
		return "ongoing"
	case capsule.JobState_JOB_STATE_COMPLETED:
		return "completed"
	case capsule.JobState_JOB_STATE_FAILED:
		return "failed"
	case capsule.JobState_JOB_STATE_TERMINATED:
		return "terminated"
	default:
		return ""
	}
}

func parseStates() ([]capsule.JobState, error) {
	if states == "" {
		return nil, nil
	}

	var res []capsule.JobState
	splits := strings.Split(states, ",")
	for _, s := range splits {
		s = strings.TrimSpace(s)
		switch s {
		case "ongoing":
			res = append(res, capsule.JobState_JOB_STATE_ONGOING)
		case "completed":
			res = append(res, capsule.JobState_JOB_STATE_COMPLETED)
		case "failed":
			res = append(res, capsule.JobState_JOB_STATE_FAILED)
		case "terminated":
			res = append(res, capsule.JobState_JOB_STATE_TERMINATED)
		default:
			return nil, fmt.Errorf("unknown state %q", s)
		}
	}

	return res, nil
}

func parseTime(s string, t *time.Time) error {
	if s == "" {
		return nil
	}
	tt, err := time.Parse(time.DateTime, s)
	if err != nil {
		return err
	}

	*t = tt

	return nil
}
