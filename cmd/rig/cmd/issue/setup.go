package issue

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/rigdev/rig-go-api/api/v1/issue"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	projectFilter     string
	environmentFilter string
	capsuleFilter     string

	limit  int
	offset int
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Scope    scope.Scope
	Prompter common.Prompter
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Scope = c.Scope
	cmd.Prompter = c.Prompter
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	issue := &cobra.Command{
		Use:               "issues",
		Short:             "List issues in Rig",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			"auth.OmitProject":     "",
			"auth.OmitEnvironment": "",
		},
		GroupID: common.ManagementGroupID,
		RunE:    cli.CtxWrap(cmd.list),
	}

	issue.Flags().IntVar(
		&limit, "limit", 10,
		"Limit the number of activities returned. Default is 10.",
	)

	issue.Flags().IntVar(
		&offset, "offset", 0,
		"Offset the activities returned. Default is 0.",
	)

	issue.Flags().StringVar(
		&projectFilter, "project-filter", "",
		"Filter activities by project ID",
	)

	issue.Flags().StringVar(
		&environmentFilter, "environment-filter", "",
		"Filter activities by environment ID",
	)

	issue.Flags().StringVar(
		&capsuleFilter, "capsule-filter", "",
		"Filter activities by capsule ID",
	)

	parent.AddCommand(issue)
}

func (c *Cmd) list(ctx context.Context, _ *cobra.Command, _ []string) error {
	resp, err := c.Rig.Issue().GetIssues(ctx, connect.NewRequest(&issue.GetIssuesRequest{
		Pagination: &model.Pagination{
			Limit:  uint32(limit),
			Offset: uint32(offset),
		},
		Filter: &issue.Filter{
			Project:     projectFilter,
			Environment: environmentFilter,
			Capsule:     capsuleFilter,
		},
	}))
	if err != nil {
		return err
	}

	issues := resp.Msg.GetIssues()

	if len(issues) == 0 {
		fmt.Println("No issues found")
		return nil
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(issues, flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{
		"Type", "Message", "Level", "Reference", "Count", "CreatedAt", "UpdatedAt", "ClosedAt", "StaleAt",
	})

	for _, i := range issues {
		t.AppendRow(table.Row{
			i.Type, i.Message, i.GetLevel().String(), issueReferenceToString(i.GetReference()),
			fmt.Sprint(i.GetCount()),
			i.GetCreatedAt().AsTime().Local().Format("2006-01-02 15:04:05"),
			i.GetUpdatedAt().AsTime().Local().Format("2006-01-02 15:04:05"),
			i.GetClosedAt().AsTime().Local().Format("2006-01-02 15:04:05"),
			i.GetStaleAt().AsTime().Local().Format("2006-01-02 15:04:05"),
		})
	}

	fmt.Println(t.Render())

	return nil
}

func issueReferenceToString(r *issue.Reference) string {
	if r == nil {
		return "all"
	}

	var str string
	if r.GetProjectId() != "" {
		str += fmt.Sprintf("Project:%s\n", r.GetProjectId())
	}

	if r.GetEnvironmentId() != "" {
		str += fmt.Sprintf("Environment:%s\n", r.GetEnvironmentId())
	}

	if r.GetCapsuleId() != "" {
		str += fmt.Sprintf("Capsule:%s\n", r.GetCapsuleId())
	}

	if r.GetRolloutId() != 0 {
		str += fmt.Sprintf("Rollout:%v\n", r.GetRolloutId())
	}

	if r.GetInstanceId() != "" {
		str += fmt.Sprintf("Instance:%s\n", r.GetInstanceId())
	}

	return str
}
