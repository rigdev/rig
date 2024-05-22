package root

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/deploy"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/image"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/instance"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/jobs"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/rollout"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/scale"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	deploymentGroupTitle      = "Deployment Commands"
	troubleshootingGroupTitle = "Troubleshooting Commands"
	basicGroupTitle           = "Basic Commands"
)

var (
	offset             int
	limit              int
	instanceID         string
	follow             bool
	previousContainers bool
	verbose            bool
	spec               bool
)

var since string

type Cmd struct {
	fx.In

	Rig      rig.Client
	Scope    scope.Scope
	Prompter common.Prompter
	Scheme   *runtime.Scheme
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	capsuleCmd := &cobra.Command{
		Use:   "capsule",
		Short: "Manage capsules",
		PersistentPreRunE: s.MakeInvokePreRunE(
			initCmd,
			func(ctx context.Context, cmd Cmd, c *cobra.Command, args []string) error {
				return cmd.persistentPreRunE(ctx, c, args)
			},
		),
		GroupID: common.CapsuleGroupID,
	}

	capsuleCmd.AddGroup(
		&cobra.Group{
			ID:    capsule.BasicGroupID,
			Title: basicGroupTitle,
		},
		&cobra.Group{
			ID:    capsule.DeploymentGroupID,
			Title: deploymentGroupTitle,
		},
		&cobra.Group{
			ID:    capsule.TroubleshootingGroupID,
			Title: troubleshootingGroupTitle,
		})

	capsuleCreate := &cobra.Command{
		Use:   "create [capsule]",
		Short: "Create a new capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.create),
		Annotations: map[string]string{
			auth.OmitCapsule:     "",
			auth.OmitEnvironment: "",
		},
		GroupID: capsule.BasicGroupID,
	}
	capsuleCmd.AddCommand(capsuleCreate)

	capsuleStop := &cobra.Command{
		Use:   "stop [capsule]",
		Short: "Stop the current rollout. This will remove all the resources related to this rollout.",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE:    cli.CtxWrap(cmd.stop),
		GroupID: capsule.DeploymentGroupID,
	}
	capsuleCmd.AddCommand(capsuleStop)

	capsuleDelete := &cobra.Command{
		Use:   "delete [capsule]",
		Short: "Delete a capsule",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
		},
		GroupID: capsule.BasicGroupID,
		RunE:    cli.CtxWrap(cmd.delete),
	}
	capsuleCmd.AddCommand(capsuleDelete)

	capsuleGet := &cobra.Command{
		Use:   "get [capsule]",
		Short: "Get a capsule",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.get),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
		},
		GroupID: capsule.BasicGroupID,
	}
	capsuleGet.Flags().BoolVarP(
		&spec, "spec", "s", false, "will display the current spec of the capsule in its environments",
	)
	capsuleCmd.AddCommand(capsuleGet)

	capsuleList := &cobra.Command{
		Use:   "list",
		Short: "List capsules",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.list),
		Annotations: map[string]string{
			auth.OmitCapsule: "",
		},
		GroupID: capsule.BasicGroupID,
	}
	capsuleList.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	capsuleList.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	capsuleCmd.AddCommand(capsuleList)

	capsuleLogs := &cobra.Command{
		Use:   "logs [capsule]",
		Short: "Get logs across all instances of the capsule",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE:    cli.CtxWrap(cmd.logs),
		GroupID: capsule.TroubleshootingGroupID,
	}
	capsuleLogs.Flags().BoolVarP(
		&follow, "follow", "f", false, "keep the connection open and read out logs as they are produced",
	)
	capsuleLogs.Flags().BoolVarP(
		&previousContainers, "previous-containers", "p", false,
		"Return logs from previous container terminations of the instance.",
	)
	capsuleLogs.Flags().StringVarP(&since, "since", "s", "", "do not show logs older than 'since'")
	capsuleCmd.AddCommand(capsuleLogs)

	capsuleStatus := &cobra.Command{
		Use:   "status [capsule]",
		Short: "Get the status of a capsule in an environment",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE:    cli.CtxWrap(cmd.status),
		GroupID: capsule.TroubleshootingGroupID,
	}
	capsuleStatus.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"show more detailed status information. Only valid with --output=pretty (default)")
	capsuleStatus.Flags().BoolVarP(&follow, "follow", "f", false,
		"keep the connection open and read status until canceled. Only valid with --verbose")
	capsuleCmd.AddCommand(capsuleStatus)

	capsulePortForward := &cobra.Command{
		Use:   "port-forward [capsule] [port]",
		Short: "Forward local request to an instance of the capsule",
		// nolint:lll
		Long: `Forward local request to an instance of the capsule.

A connection is established to an arbitrary instance (this can be overridden by the '--instance' flag).

The connection will target one of the network interfaces, and forward local traffic on the same port to this address, e.g.:

	$ rig capsule port-forward my-capsule http
	$ rig capsule port-forward my-capsule 80

Either command will forward local traffic on port 80 to the capsule.

To change the local port, you can use an override:

	$ rig capsule port-forward my-capsule 8080:http
	$ rig capsule port-forward my-capsule 8080:80

Here local traffic on port 8080 will be forwarded to port 80 of the capsule.

While the port-forwarding is running, a live log feed can be printed with the same
command, from the same instance as the connection is forwarded to:

	$ rig capsule port-forward my-capsule 80 -f
	$ rig capsule port-forward my-capsule 80 --print-logs

Finally a --verbose command can be used to troubleshoot errors related to the
forwarded traffic.
`,
		Args: cobra.MaximumNArgs(2),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE:    cli.CtxWrap(cmd.portForward),
		GroupID: capsule.TroubleshootingGroupID,
	}
	capsulePortForward.Flags().BoolVarP(
		&follow, "print-logs", "f", false, "print the instance logs while the port-forwarding is running",
	)
	capsulePortForward.Flags().BoolVarP(
		&verbose, "verbose", "v", false, "print verbose information about the connections forwarded",
	)
	capsulePortForward.Flags().StringVar(
		&instanceID, "instance", "", "a specific instance to connect to",
	)
	capsuleCmd.AddCommand(capsulePortForward)

	capsuleUpdate := &cobra.Command{
		Use:   "update",
		Short: "Update the settings of the capsule",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.update),
	}
	capsuleCmd.AddCommand(capsuleUpdate)

	capsuleListProposal := &cobra.Command{
		Use:   "list-proposals",
		Short: "Lists the ongoing capsule rollout proposals",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.listProposals),
	}
	capsuleCmd.AddCommand(capsuleListProposal)

	parent.AddCommand(capsuleCmd)

	scale.Setup(capsuleCmd, s)
	image.Setup(capsuleCmd, s)
	deploy.Setup(capsuleCmd, s)
	instance.Setup(capsuleCmd, s)
	rollout.Setup(capsuleCmd, s)
	jobs.Setup(capsuleCmd, s)
}

func (c *Cmd) completions(
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

func (c *Cmd) persistentPreRunE(ctx context.Context, cmd *cobra.Command, args []string) error {
	if _, ok := cmd.Annotations[auth.OmitCapsule]; ok {
		return nil
	}

	if len(args) > 0 {
		capsule.CapsuleID = args[0]
		return nil
	}

	name, err := capsule.SelectCapsule(ctx, c.Rig, c.Prompter, c.Scope)
	if err != nil {
		return err
	}

	capsule.CapsuleID = name
	return nil
}
