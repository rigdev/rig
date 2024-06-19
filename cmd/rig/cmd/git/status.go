package git

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) status(ctx context.Context, _ *cobra.Command, _ []string) error {
	resp, err := c.Rig.Settings().GetGitStoreStatus(ctx, connect.NewRequest(&settings.GetGitStoreStatusRequest{}))
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg, flags.Flags.OutputType)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	tbl := table.New(
		"Repository", "| Branch", "| Last Successful Commit", "| Age", "| Last Processed Commit", "| Age", "| Error",
	).
		WithPadding(1).
		WithHeaderFormatter(headerFmt)
	for _, repo := range resp.Msg.GetRepositories() {
		rb := repo.GetRepo()
		status := repo.GetStatus()
		row := append([]any{rb.GetRepository(), rb.GetBranch()}, formatStatusRow(status)...)
		tbl.AddRow(row...)
	}
	tbl.Print()

	fmt.Println("")

	tbl = table.New(
		"Project", "| Capsule", "| Environment",
		"| Last Successful Commt", "| Age", "| Last Processed Commit",
		"| Age", "| Error",
	)
	tbl.WithHeaderFormatter(headerFmt)
	for _, capsule := range resp.Msg.GetCapsules() {
		status := capsule.GetStatus()
		id := capsule.GetCapsule()
		row := append(
			[]any{id.GetProject(), id.GetCapsule(), id.GetEnvironment()},
			formatStatusRow(status)...,
		)
		tbl.AddRow(row...)
	}
	tbl.Print()

	fmt.Println("")

	if len(resp.Msg.GetErrors()) > 0 {
		tbl = table.New("Git Webhook Error", "| Age").
			WithPadding(1).
			WithHeaderFormatter(headerFmt)
		for _, err := range resp.Msg.GetErrors() {
			tbl.AddRow(err.GetErr(), formatTime(err.GetTimestamp().AsTime()))
		}
		tbl.Print()
	}

	return nil
}

func formatStatusRow(status *model.GitStatus) []any {
	return []any{
		formatCommit(status.GetLastSuccessfulCommitId()), formatTime(status.GetLastSuccessfulCommitTime().AsTime()),
		formatCommit(status.GetLastProcessedCommitId()), formatTime(status.GetLastProcessedCommitTime().AsTime()),
		common.StringOr(status.GetError(), "-"),
	}
}

func formatCommit(commit string) string {
	if commit == "" {
		return "-"
	}
	return commit[:12]
}

func formatTime(t time.Time) string {
	if t.IsZero() || t.Unix() == 0 {
		return "-"
	}
	d := time.Since(t)
	return common.FormatDuration(d)
	// return fmt.Sprintf("%s ago", common.FormatDuration(d))
}
