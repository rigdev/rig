package pipeline

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	pipeline_api "github.com/rigdev/rig-go-api/api/v1/capsule/pipeline"
	project_api "github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Scope    scope.Scope
	Prompter common.Prompter
	Scheme   *runtime.Scheme
}

var (
	offset       int
	limit        int
	pipelineName string
	dryRun       bool
	force        bool
)

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	pipeline := &cobra.Command{
		Use:               "pipeline",
		Short:             "Manage and view pipeline executions",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		GroupID:           capsule.DeploymentGroupID,
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
		},
	}

	pipelineList := &cobra.Command{
		Use:   "list [capsule]",
		Short: "List pipeline executions",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
		RunE: cli.CtxWrap(cmd.list),
	}
	pipelineList.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	pipelineList.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	pipelineList.Flags().StringVar(&pipelineName, "pipeline", "", "filter by pipeline name")
	if err := pipelineList.RegisterFlagCompletionFunc("pipeline",
		cli.HackCtxWrapCompletion(cmd.pipelineNameCompletion, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pipeline.AddCommand(pipelineList)

	pipelineGet := &cobra.Command{
		Use:   "get [execution-id]",
		Short: "Get pipeline execution details",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.pipelineStatusCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
		Annotations: map[string]string{
			auth.OmitCapsule: "",
		},
		RunE: cli.CtxWrap(cmd.get),
	}
	pipeline.AddCommand(pipelineGet)

	pipelineStart := &cobra.Command{
		Use:   "start [capsule-id] [pipeline-name]",
		Short: "Start a pipeline execution on a capsule.",
		Args:  cobra.MaximumNArgs(2),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.capsuleCompletions, s),
			cli.HackCtxWrapCompletion(cmd.pipelineNameCompletion, s),
		),
		RunE: cli.CtxWrap(cmd.start),
	}
	pipeline.AddCommand(pipelineStart)

	pipelineAbort := &cobra.Command{
		Use: "abort [execution-id]",
		Short: "Abort a pipeline execution ." +
			"This will not stop any capsules running, but prevent any further progressions from being made.",
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.pipelineStatusCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
		RunE: cli.CtxWrap(cmd.abort),
		Annotations: map[string]string{
			auth.OmitCapsule: "",
		},
	}
	pipeline.AddCommand(pipelineAbort)

	pipelineProgress := &cobra.Command{
		Use: "promote [execution-id]",
		Short: "promote the pipeline to the next phase. " +
			"This will only work if the pipeline is in a state that allows promotion. " +
			"I.e. it has a manual trigger",
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.pipelineStatusCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
		RunE: cli.CtxWrap(cmd.progress),
		Annotations: map[string]string{
			auth.OmitCapsule: "",
		},
	}
	pipelineProgress.Flags().BoolVar(&dryRun, "dry-run", false,
		"Dry run the promotion. If interactive, it will interactively show the diffs. "+
			"Otherwise it will print the resulting resources.")
	pipelineProgress.Flags().BoolVar(&force, "force", false,
		"Force the promotion. This will bypass any ready checks and "+
			"force a manual promotion no matter the configured triggers.")
	pipeline.AddCommand(pipelineProgress)

	parent.AddCommand(pipeline)
}

func (c *Cmd) capsuleCompletions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions.Capsules(ctx, c.Rig, toComplete, c.Scope)
}

func (c *Cmd) promptForPipelineName(ctx context.Context) (string, error) {
	resp, err := c.Rig.Project().GetEffectivePipelineSettings(ctx,
		connect.NewRequest(&project_api.GetEffectivePipelineSettingsRequest{
			ProjectId: c.Scope.GetCurrentContext().GetProject(),
			CapsuleId: capsule_cmd.CapsuleID,
		}))
	if err != nil {
		return "", err
	}

	header := []string{
		"Name",
		"1stEnv",
		"#Phases",
		"Running",
	}

	var rows [][]string
	for _, pipeline := range resp.Msg.GetPipelines() {
		rows = append(rows, []string{
			pipeline.GetPipeline().GetName(),
			pipeline.GetPipeline().GetInitialEnvironment(),
			fmt.Sprint(len(pipeline.GetPipeline().GetPhases())),
			fmt.Sprint(pipeline.GetAlreadyRunning()),
		})
	}

	i, err := c.Prompter.TableSelect("Select a pipeline", rows, header)
	if err != nil {
		return "", err
	}

	return resp.Msg.GetPipelines()[i].GetPipeline().GetName(), nil
}

func (c *Cmd) promptForPipelineID(ctx context.Context) (string, error) {
	resp, err := c.Rig.Capsule().ListPipelineStatuses(ctx,
		connect.NewRequest(&capsule_api.ListPipelineStatusesRequest{
			ProjectFilter: c.Scope.GetCurrentContext().GetProject(),
			Pagination: &model.Pagination{
				Descending: true,
			},
		}))
	if err != nil {
		return "", err
	}

	var rows [][]string
	for _, pipeline := range resp.Msg.GetStatuses() {
		rows = append(rows, []string{
			fmt.Sprint(pipeline.GetExecutionId()),
			pipeline.GetCapsuleId(),
			pipeline.GetPipelineName(),
			pipeline.GetState().String(),
		})
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("no pipeline executions found")
	}

	i, err := c.Prompter.TableSelect("Pipeline ID", rows, []string{
		"ID",
		"Capsule",
		"Pipeline",
		"State",
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprint(resp.Msg.Statuses[i].GetExecutionId()), nil
}

func (c *Cmd) pipelineStatusCompletions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Capsule().ListPipelineStatuses(ctx,
		connect.NewRequest(&capsule_api.ListPipelineStatusesRequest{
			ProjectFilter: c.Scope.GetCurrentContext().GetProject(),
			Pagination: &model.Pagination{
				Descending: true,
			},
		}))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var pipelineStatusIDs []string
	for _, status := range resp.Msg.GetStatuses() {
		if strings.HasPrefix(fmt.Sprint(status.GetExecutionId()), toComplete) {
			pipelineStatusIDs = append(pipelineStatusIDs, formatPipelineStatus(status))
		}
	}

	if len(pipelineStatusIDs) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return pipelineStatusIDs, cobra.ShellCompDirectiveDefault
}

func formatPipelineStatus(status *pipeline_api.Status) string {
	var startedAt string
	if status.GetStartedAt().AsTime().IsZero() {
		startedAt = "-"
	} else {
		startedAt = time.Since(status.GetStartedAt().AsTime()).Truncate(time.Second).String()
	}
	return fmt.Sprintf("%d\t (Capsule: %s, Pipeline %s, State: %v, Started At: %v)",
		status.GetExecutionId(), status.GetCapsuleId(), status.GetPipelineName(), status.GetState(), startedAt)
}

func (c *Cmd) pipelineNameCompletion(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Project().GetEffectivePipelineSettings(ctx,
		connect.NewRequest(&project_api.GetEffectivePipelineSettingsRequest{
			ProjectId: c.Scope.GetCurrentContext().GetProject(),
		}))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var pipelineNames []string
	for _, pipeline := range resp.Msg.GetPipelines() {
		if strings.HasPrefix(pipeline.GetPipeline().GetName(), toComplete) {
			pipelineNames = append(pipelineNames, pipeline.GetPipeline().GetName())
		}
	}

	if len(pipelineNames) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return pipelineNames, cobra.ShellCompDirectiveDefault
}
