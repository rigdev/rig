package capsule

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset   int
	limit    int
	rollout  uint64
	replicas int
)

var (
	deploy      bool
	follow      bool
	interactive bool
	outputJSON  bool
)

var (
	name        string
	image       string
	buildID     string
	networkFile string
	instanceID  string
)

func Setup(parent *cobra.Command) {
	capsule := &cobra.Command{
		Use: "capsule",
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a new capsule",
		Args:  cobra.NoArgs,
		RunE:  base.Register(CapsuleCreate),
	}
	create.Flags().StringVarP(&name, "name", "n", "", "name of the capsule")
	create.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive mode")

	capsule.AddCommand(create)

	push := &cobra.Command{
		Use:   "push [capsule-name]",
		Short: "Push a local image to a Capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsulePush),
	}
	push.Flags().StringVarP(&image, "image", "i", "", "image to push")
	push.Flags().BoolVarP(&deploy, "deploy", "d", false, "deploy build after successful push")
	capsule.AddCommand(push)

	createBuild := &cobra.Command{
		Use:   "create-build [capsule-name]",
		Short: "Create a new build with the given image",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleCreateBuild),
	}
	createBuild.Flags().StringVarP(&image, "image", "i", "", "image to use for the build")
	createBuild.Flags().BoolVarP(&deploy, "deploy", "d", false, "deploy build after successful creation")
	capsule.AddCommand(createBuild)

	deploy := &cobra.Command{
		Use:   "deploy [capsule-name]",
		Short: "Deploy the given build to a capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleDeploy),
	}
	deploy.Flags().StringVarP(&buildID, "build-id", "b", "", "build id to deploy")
	capsule.AddCommand(deploy)

	scale := &cobra.Command{
		Use:   "scale [capsule-name]",
		Short: "scale the capsule to a new number of replicas",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleScale),
	}

	scale.Flags().IntVarP(&replicas, "replicas", "r", -1, "number of replicas to scale to")
	capsule.AddCommand(scale)

	abort := &cobra.Command{
		Use:   "abort [capsule-name]",
		Short: "abort the current rollout. This will leave the capsule in a undefined state",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleAbort),
	}
	capsule.AddCommand(abort)

	configureNetwork := &cobra.Command{
		Use:   "configure-network [capsule-name]",
		Short: "configure the network of the capsule",
		Args:  cobra.MaximumNArgs(2),
		RunE:  base.Register(CapsuleConfigureNetwork),
	}
	configureNetwork.Flags().StringVarP(&networkFile, "network-file", "n", "", "network file to use")
	capsule.AddCommand(configureNetwork)

	delete := &cobra.Command{
		Use:   "delete [capsule-name]",
		Short: "Delete a capsule",
		Args:  cobra.ExactArgs(1),
		RunE:  base.Register(CapsuleDelete),
	}
	capsule.AddCommand(delete)

	list := &cobra.Command{
		Use:     "list",
		Short:   "List capsules",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE:    base.Register(CapsuleList),
	}
	list.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	list.Flags().IntVarP(&offset, "offset", "o", 0, "offset for pagination")
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	capsule.AddCommand(list)

	listBuilds := &cobra.Command{
		Use:   "list-builds [capsule-name]",
		Short: "List builds",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleListBuilds),
	}
	listBuilds.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	listBuilds.Flags().IntVarP(&offset, "offset", "o", 0, "offset for pagination")
	listBuilds.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	capsule.AddCommand(listBuilds)

	listRollouts := &cobra.Command{
		Use:   "list-rollouts [capsule-name]",
		Short: "List rollouts",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleListRollouts),
	}
	listRollouts.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	listRollouts.Flags().IntVarP(&offset, "offset", "o", 0, "offset for pagination")
	listRollouts.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	capsule.AddCommand(listRollouts)

	listInstances := &cobra.Command{
		Use:   "list-instances [capsule-name]",
		Short: "List instances",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleListInstances),
	}
	listInstances.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	listInstances.Flags().IntVarP(&offset, "offset", "o", 0, "offset for pagination")
	listInstances.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	capsule.AddCommand(listInstances)

	restartInstance := &cobra.Command{
		Use:   "restart-instance [capsule-name]",
		Short: "Restart a single instance",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleRestartInstance),
	}
	restartInstance.Flags().StringVarP(&instanceID, "instance-id", "i", "", "instance id to restart")
	capsule.AddCommand(restartInstance)

	logs := &cobra.Command{
		Use:   "logs [capsule-name]",
		Short: "Read instance logs from the capsule ",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleLogs),
	}
	logs.Flags().StringVarP(&instanceID, "instance-id", "i", "", "instance id to restart")
	logs.Flags().BoolVarP(&follow, "follow", "f", false, "keep the connection open and read out logs as they are produced")
	capsule.AddCommand(logs)

	config := &cobra.Command{
		Use:   "config [capsule-name]",
		Short: "Configure a capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleConfig),
	}
	config.Flags().BoolP("auto-add-service-account", "a", false, "automatically create and add Rig service-account for this capsule")
	capsule.AddCommand(config)

	events := &cobra.Command{
		Use:   "events [capsule-name]",
		Short: "List events related to a rollout, default to the current rollout",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(CapsuleEvents),
	}
	events.Flags().Uint64VarP(&rollout, "rollout", "r", 0, "rollout to get events from")
	capsule.AddCommand(events)

	setupSetResources(capsule)
	setupGetResources(capsule)

	parent.AddCommand(capsule)

	base.AddOptions(
		fx.Provide(
			provideCapsuleID,
		),
	)
}

type CapsuleID = uuid.UUID

func provideCapsuleID(ctx context.Context, nc rig.Client, args []string) (CapsuleID, error) {
	var capsuleName string
	var err error
	if len(args) == 0 {
		capsuleName, err = common.PromptInput("Enter Capsule name:", common.ValidateNonEmptyOpt)
		if err != nil {
			return "", err
		}
	} else {
		capsuleName = args[0]
	}
	res, err := nc.Capsule().GetByName(ctx, &connect.Request[capsule.GetByNameRequest]{
		Msg: &capsule.GetByNameRequest{
			Name: capsuleName,
		},
	})
	if err != nil {
		return "", err
	}

	return uuid.UUID(res.Msg.GetCapsule().GetCapsuleId()), nil
}

type InstanceID = string

func provideInstanceID(ctx context.Context, nc rig.Client, capsuleID CapsuleID, arg string) (InstanceID, error) {
	if arg != "" {
		return arg, nil
	}

	res, err := nc.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
		Msg: &capsule.ListInstancesRequest{
			CapsuleId: capsuleID.String(),
		},
	})
	if err != nil {
		return "", err
	}

	var items []string
	for _, i := range res.Msg.GetInstances() {
		items = append(items, i.GetInstanceId())
	}

	if len(items) == 0 {
		return "", errors.InvalidArgumentErrorf("no instances selected")
	}

	if len(items) == 1 {
		return items[0], nil
	}

	_, s, err := common.PromptSelect("instance", items)
	return s, err
}
