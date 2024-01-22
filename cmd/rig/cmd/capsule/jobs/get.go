package jobs

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

const maxColLength = 20

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, _ []string) error {
	rollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if errors.IsNotFound(err) {
		cmd.Println("No cron jobs set")
		return nil
	} else if err != nil {
		return err
	}

	cronJobs := rollout.GetConfig().GetCronJobs()
	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(cronJobs, flags.Flags.OutputType)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Name", "Schedule", "Command", "Timeout", "Max Retries")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, j := range cronJobs {
		timeoutString := "-"
		if t := j.GetTimeout(); t != nil {
			timeoutString = t.AsDuration().String()
		}
		row := formatRow([]any{
			j.GetJobName(),
			j.GetSchedule(),
			formatCommand(j),
			timeoutString,
			j.GetMaxRetries(),
		})
		tbl.AddRow(row...)
	}
	tbl.Print()

	return nil
}

func formatCommand(j *capsule.CronJob) string {
	if url := j.GetUrl(); url != nil {
		return fmt.Sprintf(":%v%s", url.GetPort(), url.GetPath())
	}
	if cmd := j.GetCommand(); cmd != nil {
		return cmd.GetCommand() + " " + strings.Join(cmd.GetArgs(), " ")
	}
	return ""
}

func formatRow(row []any) []any {
	var res []any
	for _, r := range row {
		s := fmt.Sprint(r)
		if len(s) > maxColLength-3 {
			s = s[:maxColLength-3] + "..."
		}
		res = append(res, s)
	}
	return res
}
