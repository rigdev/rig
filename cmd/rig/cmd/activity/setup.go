package activity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/activity"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	fromStr string
	toStr   string
	since   string
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
	activity := &cobra.Command{
		Use:               "activities",
		Short:             "List activities in Rig",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			"auth.OmitProject":     "",
			"auth.OmitEnvironment": "",
		},
		GroupID: common.ManagementGroupID,
		RunE:    cli.CtxWrap(cmd.list),
	}

	activity.Flags().StringVarP(
		&fromStr, "from", "f", "",
		"If set, only include activities after this date. Layout is 2006-01-02 15:04:05. Default is 24 hours ago",
	)
	activity.Flags().StringVarP(
		&toStr, "to", "t", "",
		"If set, only include activites before this date. Layout is 2006-01-02 15:04:05. Default is now.",
	)
	activity.Flags().StringVarP(
		&since, "since", "s", "",
		"A duration. If set, only include activities younger than 'since'. "+
			"Cannot be used if either --from or --to is used. Default is 24 hours.",
	)

	parent.AddCommand(activity)
}

func (c *Cmd) list(ctx context.Context, _ *cobra.Command, _ []string) error {
	from, to, err := parseFromTo()
	if err != nil {
		return err
	}

	resp, err := c.Rig.Activity().GetActivities(ctx, connect.NewRequest(&activity.GetActivitiesRequest{
		From: timestamppb.New(from),
		To:   timestamppb.New(to),
	}))
	if err != nil {
		return err
	}

	activities := resp.Msg.GetActivities()

	if len(activities) == 0 {
		fmt.Println("No activities found")
		return nil
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(activities, flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{
		"Topic", "Message", "Scope", "Timestamp",
	})

	for _, a := range activities {
		t.AppendRow(table.Row{
			common.TopicToString(a.GetTopic()), a.GetMessage(), activityScopeToString(a.GetScope()),
			a.GetTimestamp().AsTime().Local().Format("2006-01-02 15:04:05"),
		})
	}

	fmt.Println(t.Render())

	return nil
}

func activityScopeToString(s *activity.Scope) string {
	if s == nil {
		return "All"
	}

	str := ""
	if s.GetProject() != "" {
		str += fmt.Sprintf("Project: %s\n", s.GetProject())
	}

	if s.GetEnvironment() != "" {
		str += fmt.Sprintf("Environment: %s\n", s.GetEnvironment())
	}

	if s.GetCapsule() != "" {
		str += fmt.Sprintf("Capsule: %s\n", s.GetCapsule())
	}

	if s.GetUser() != "" {
		str += fmt.Sprintf("User: %s\n", s.GetUser())
	}

	return str
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
