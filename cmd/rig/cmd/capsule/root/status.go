package root

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/capsule/rollout"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/scale"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

var red = color.New(color.FgRed)
var green = color.New(color.FgGreen)
var blue = color.New(color.FgBlue)
var yellow = color.New(color.FgYellow)

func (c *Cmd) status(ctx context.Context, _ *cobra.Command, _ []string) error {
	if !verbose {
		statusResp, err := c.Rig.Capsule().GetStatus(ctx, &connect.Request[capsule.GetStatusRequest]{
			Msg: &capsule.GetStatusRequest{
				CapsuleId:     capsule_cmd.CapsuleID,
				ProjectId:     flags.GetProject(c.Scope),
				EnvironmentId: flags.GetEnvironment(c.Scope),
			},
		})
		if err != nil {
			return err
		}

		status := statusResp.Msg.GetStatus()
		if flags.Flags.OutputType != common.OutputTypePretty {
			return common.FormatPrint(status, flags.Flags.OutputType)
		}

		rolloutResp, err := c.Rig.Capsule().GetRollout(ctx, connect.NewRequest(&capsule.GetRolloutRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			RolloutId: status.GetCurrentRolloutId(),
			ProjectId: flags.GetProject(c.Scope),
		}))
		if err != nil {
			return err
		}

		summary := buildStatusSummary(status, rolloutResp.Msg.GetRollout())
		fmt.Print(summary)
		return nil
	}

	return showStatus(ctx, c, capsule_cmd.CapsuleID)
}

// -------------------- Summary --------------------

func buildStatusSummary(s *capsule.Status, r *capsule.Rollout) string {
	builder := &strings.Builder{}
	buildCapsuleInfo(builder, s)
	buildRolloutStatus(builder, r)
	buildContainerConfig(builder, s.GetContainerConfig())
	buildInstanceStatus(builder, s.GetInstances())
	buildConfigFileStatus(builder, s.GetConfigFiles())
	buildInterfaceStatus(builder, s.GetInterfaces())
	buildCronjobStatus(builder, s.GetCronJobs())
	return builder.String()
}

func buildCapsuleInfo(builder *strings.Builder, s *capsule.Status) {
	builder.WriteString(fmt.Sprintf("Namespace: %s\n", s.GetNamespace()))
	builder.WriteString(fmt.Sprintf("Current Rollout: %d\n", s.GetCurrentRolloutId()))
}

func buildRolloutStatus(builder *strings.Builder, ro *capsule.Rollout) {
	builder.WriteString(fmt.Sprintf("Rollout %d\n", ro.GetRolloutId()))
	createdAt := ro.GetConfig().GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")
	r := ro.GetStatus()

	currentStageIcon := ""
	switch r.GetState() {
	case rollout.State_STATE_CONFIGURE:
		currentStageIcon = StageStateToIcon(r.GetStages().GetConfigure().GetInfo().GetState())
	case rollout.State_STATE_RESOURCE_CREATION:
		currentStageIcon = StageStateToIcon(r.GetStages().GetResourceCreation().GetInfo().GetState())
	case rollout.State_STATE_RUNNING:
		currentStageIcon = StageStateToIcon(r.GetStages().GetRunning().GetInfo().GetState())
	}
	builder.WriteString(GetIndented(fmt.Sprintf("Current Stage: %s %s",
		RolloutStageToString(r.GetState()), currentStageIcon), 2))

	builder.WriteString(GetIndented("All Stages:", 2))

	if configure := r.GetStages().GetConfigure(); configure != nil {
		var lastStepInfo *rollout.StepInfo
		if len(configure.GetSteps()) > 0 {
			lastStepInfo = GetStepInfo(configure.GetSteps()[len(configure.GetSteps())-1])
		}
		builder.WriteString(GetIndented(StageToString("Configuring", lastStepInfo), 4))
	}
	if creation := r.GetStages().GetResourceCreation(); creation != nil {
		var lastStepInfo *rollout.StepInfo
		if len(creation.GetSteps()) > 0 {
			lastStepInfo = GetStepInfo(creation.GetSteps()[len(creation.GetSteps())-1])
		}
		builder.WriteString(GetIndented(StageToString("Resource Creation", lastStepInfo), 4))
	}
	if running := r.GetStages().GetRunning(); running != nil {
		var lastStepInfo *rollout.StepInfo
		if len(running.GetSteps()) > 0 {
			lastStepInfo = GetStepInfo(running.GetSteps()[len(running.GetSteps())-1])
		}
		builder.WriteString(GetIndented(StageToString("Running", lastStepInfo), 4))
	}

	if configure := r.GetStages().GetConfigure(); configure != nil {
		for _, s := range configure.GetSteps() {
			if commitStep := s.GetCommit(); commitStep != nil {
				builder.WriteString(GetIndented(fmt.Sprintf("Commit: %s", commitStep.GetCommitHash()), 2))
				builder.WriteString(GetIndented(fmt.Sprintf("Commit URL: %s", commitStep.GetCommitUrl()), 2))
			}
		}
	}

	if author := ro.GetConfig().GetCreatedBy(); author != nil {
		builder.WriteString(GetIndented(fmt.Sprintf("Created by: %s", author.GetPrintableName()), 2))
	}
	if !ro.GetConfig().GetCreatedAt().AsTime().IsZero() {
		builder.WriteString(GetIndented(fmt.Sprintf("Created at: %s", createdAt), 2))
	}
}

func buildContainerConfig(builder *strings.Builder, c *capsule.ContainerConfig) {
	builder.WriteString("Container Config\n")

	truncatedImg := c.GetImage()
	if len(truncatedImg) > 50 {
		truncatedImg = truncatedImg[:50] + "..."
	}
	builder.WriteString(GetIndented(fmt.Sprintf("Image: %s", truncatedImg), 2))
	if c.GetCommand() != "" {
		builder.WriteString(GetIndented(fmt.Sprintf("%s %s", c.GetCommand(), strings.Join(c.GetArgs(), " ")), 2))
	}
	if len(c.GetEnvironmentVariables()) > 0 {
		builder.WriteString(GetIndented("Environment Variables", 2))
		for key, value := range c.GetEnvironmentVariables() {
			builder.WriteString(GetIndented(fmt.Sprintf("%s=%s", key, value), 4))
		}
	}

	if c.GetScale().GetCpuTarget() == nil {
		builder.WriteString(GetIndented(fmt.Sprintf("#Replicas: %d", c.GetScale().GetMinReplicas()), 2))
	} else {
		builder.WriteString(GetIndented(fmt.Sprintf("Auto-scaling: %d-%d",
			c.GetScale().GetMinReplicas(), c.GetScale().GetMaxReplicas()), 2))
	}

	if c.GetResources().GetRequests() != nil {
		builder.WriteString(GetIndented("Requsted Resources:", 2))

		cpu := "-"
		memory := "-"
		if c.GetResources().GetRequests().GetCpuMillis() != 0 {
			cpu = scale.MilliIntToString(uint64(c.GetResources().GetRequests().GetCpuMillis()))
		}
		if c.GetResources().GetRequests().GetMemoryBytes() != 0 {
			memory = scale.IntToByteString(c.GetResources().GetRequests().GetMemoryBytes())
		}
		builder.WriteString(GetIndented(fmt.Sprintf("CPU: %s", cpu), 4))
		builder.WriteString(GetIndented(fmt.Sprintf("Memory: %s", memory), 4))

	}
}

func buildInstanceStatus(builder *strings.Builder, i *capsule.InstancesStatus) {
	builder.WriteString("Running Instances\n")
	builder.WriteString(GetIndented(green.Sprintf("#Healthy: %d", i.GetNumReady()), 2))
	builder.WriteString(GetIndented(blue.Sprintf("#Upgrading: %d", i.GetNumUpgrading()), 2))
	builder.WriteString(GetIndented(yellow.Sprintf("#Old Version: %d", i.GetNumWrongVersion()), 2))
	builder.WriteString(GetIndented(red.Sprintf("#Failing: %d", i.GetNumStuck()), 2))
}

func buildConfigFileStatus(builder *strings.Builder, c []*capsule.ConfigFileStatus) {
	if len(c) == 0 {
		builder.WriteString("No Config Files\n")
		return
	}

	builder.WriteString("Config Files\n")
	for _, cf := range c {
		transition := TransitionToIcon(cf.GetTransition())
		state := StateToIcon(getAggregatedStatus(cf.GetStatus()))
		builder.WriteString(GetIndented(fmt.Sprintf("%s %s%s", cf.GetPath(), transition, state), 2))
	}
}

func buildInterfaceStatus(builder *strings.Builder, c []*capsule.InterfaceStatus) {
	if len(c) == 0 {
		builder.WriteString("No Interfaces\n")
		return
	}
	builder.WriteString("Interfaces\n")
	for _, i := range c {
		transition := TransitionToIcon(i.GetTransition())
		state := StateToIcon(getAggregatedStatus(i.GetStatus()))
		builder.WriteString(GetIndented(fmt.Sprintf("%s:%d %s%s", i.GetName(), i.GetPort(), transition, state), 2))
		for _, r := range i.GetRoutes() {
			state := StateToIcon(getAggregatedStatus(r.GetStatus()))
			transition := TransitionToIcon(r.GetTransition())
			builder.WriteString(GetIndented(fmt.Sprintf("%s %s%s", r.GetRoute().GetHost(), transition, state), 4))
		}
	}
}

func buildCronjobStatus(builder *strings.Builder, cs []*capsule.CronJobStatus) {
	if len(cs) == 0 {
		builder.WriteString("No Cron Jobs\n")
		return
	}
	builder.WriteString("Cron Jobs\n")
	for _, c := range cs {
		transition := TransitionToIcon(c.GetTransition())
		state := StateToIcon(c.GetLastExecution())
		builder.WriteString(GetIndented(fmt.Sprintf("%s %s%s", c.GetJobName(), transition, state), 2))
	}
}

// -------------------- Verbose --------------------

func showStatus(
	ctx context.Context,
	c *Cmd,
	capsuleID string,
) error {
	ctx, cancel := context.WithCancel(ctx)
	var s *capsule.Status
	rolloutID := uint64(0)
	errChan := make(chan error, 1)
	statusChan := make(chan *capsule.Status, 1)
	rolloutChan := make(chan *capsule.Rollout, 1)
	var rolloutCancelFunc context.CancelFunc
	var rolloutCtx context.Context

	defer func() {
		cancel()
		close(errChan)
		close(statusChan)
		close(rolloutChan)
		if rolloutCancelFunc != nil {
			rolloutCancelFunc()
		}
	}()

	var objectTree *tview.TreeView
	var rolloutSummary, instanceSummary, selectedView *tview.TextView

	grid := tview.NewGrid().
		SetRows(-1, -1).
		SetColumns(-1, -2, -2)

	app := tview.NewApplication().SetRoot(grid, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if app.GetFocus() != objectTree {
				return event
			}

			text, state := getSelectedObjectTextAndState(objectTree.GetCurrentNode().GetText(), s)

			selectedView.SetTitle(objectTree.GetCurrentNode().GetText())
			selectedView.SetTitleColor(StateToTCellColor(state))
			selectedView.SetBorderColor(StateToTCellColor(state))
			selectedView.SetText(tview.TranslateANSI(text))
			event = nil
		case tcell.KeyTab:
			// shift focus to the next item
			if app.GetFocus() == objectTree {
				app.SetFocus(selectedView)
			} else if app.GetFocus() == selectedView {
				app.SetFocus(instanceSummary)
			} else if app.GetFocus() == instanceSummary {
				app.SetFocus(rolloutSummary)
			} else if app.GetFocus() == rolloutSummary {
				app.SetFocus(objectTree)
			}
		case tcell.KeyEsc, tcell.KeyCtrlC:
			cancel()
		}
		return event
	})

	go watchCapsuleStatus(ctx, c, capsuleID, errChan, statusChan)
	go func() {
		err := app.Run()
		if err != nil {
			errChan <- err
		}
	}()

	defer app.Stop()

	for {
		select {
		case err := <-errChan:
			return err
		case s = <-statusChan:
			app.QueueUpdateDraw(func() {
				if rolloutID != s.GetCurrentRolloutId() {
					rolloutID = s.GetCurrentRolloutId()
					if rolloutCancelFunc != nil {
						rolloutCancelFunc()
					}
					rolloutCtx, rolloutCancelFunc = context.WithCancel(ctx)
					go watchRollout(rolloutCtx, c, capsuleID, rolloutID, errChan, rolloutChan)
				}

				currentlySelected := ""
				if objectTree != nil {
					currentlySelected = objectTree.GetCurrentNode().GetText()
				}

				grid.RemoveItem(instanceSummary).
					RemoveItem(objectTree).
					RemoveItem(selectedView)

				instanceSummary = buildInstanceSummary(s)
				objectTree = buildObjectTree(s, capsuleID, currentlySelected)
				selectedView = buildSelectedView(objectTree.GetCurrentNode(), s)

				grid.AddItem(instanceSummary, 0, 0, 2, 1, 0, 0, false).
					AddItem(objectTree, 1, 1, 1, 1, 0, 0, true).
					AddItem(selectedView, 0, 2, 2, 1, 0, 0, false)

				app.SetFocus(objectTree)
			})
		case r := <-rolloutChan:
			app.QueueUpdateDraw(func() {
				grid.RemoveItem(rolloutSummary)
				rolloutSummary = buildRolloutSummary(r)
				grid.AddItem(rolloutSummary, 0, 1, 1, 1, 0, 0, false)
			})

		case <-ctx.Done():
			return nil
		}
	}
}

func buildRolloutSummary(ro *capsule.Rollout) *tview.TextView {
	r := ro.GetStatus()

	r.GetState()

	rolloutSummary := &strings.Builder{}
	buildRolloutStatus(rolloutSummary, ro)

	color := tcell.ColorWhite
	switch r.GetState() {
	case rollout.State_STATE_CONFIGURE:
		color = StageStateToColor(r.GetStages().GetConfigure().GetInfo().GetState())
	case rollout.State_STATE_RESOURCE_CREATION:
		color = StageStateToColor(r.GetStages().GetResourceCreation().GetInfo().GetState())
	case rollout.State_STATE_RUNNING:
		color = StageStateToColor(r.GetStages().GetRunning().GetInfo().GetState())
	}

	summary := tview.NewTextView()
	summary.SetTitle("Rollout Summary (ESC to exit)")
	summary.SetBorder(true)
	summary.SetTitleColor(color)
	summary.SetBorderColor(color)
	summary.SetDynamicColors(true)
	summary.SetText(tview.TranslateANSI(rolloutSummary.String()))
	return summary
}

func buildInstanceSummary(s *capsule.Status) *tview.TextView {
	instanceSummary := &strings.Builder{}
	buildContainerConfig(instanceSummary, s.GetContainerConfig())
	instanceSummary.WriteString("\n")
	buildInstanceStatus(instanceSummary, s.GetInstances())

	summary := tview.NewTextView()
	summary.SetTitle("Container Config And Instance Summary (ESC to exit)")
	summary.SetBorder(true)
	summary.SetDynamicColors(true)
	summary.SetText(tview.TranslateANSI(instanceSummary.String()))
	return summary
}

func buildObjectTree(s *capsule.Status, capsuleID string, selectedNodeName string) *tview.TreeView {
	var selectedNode *tview.TreeNode

	add := func(parent *tview.TreeNode, kind string, name string, transition string, status string) *tview.TreeNode {
		nodeKindName := fmt.Sprintf("%s/%s", kind, name)
		nodeName := fmt.Sprintf("%s %s%s", nodeKindName, transition, status)
		node := tview.NewTreeNode(nodeName).
			SetSelectable(true)

		if strings.HasPrefix(selectedNodeName, nodeKindName) {
			selectedNode = node
		}

		parent.AddChild(node)
		return node
	}

	root := tview.NewTreeNode(fmt.Sprintf("Capsule/%s %s", capsuleID,
		StateToIcon(getAggregatedStatus(s.Capsule.GetStatuses())))).SetSelectable(true)

	tree := tview.NewTreeView().
		SetRoot(root)

	for _, i := range s.Interfaces {
		interfaceNode := add(root, "Interface", i.GetName(), TransitionToIcon(i.GetTransition()),
			StateToIcon(getAggregatedStatus(i.GetStatus())))
		for _, r := range i.Routes {
			add(interfaceNode, "Route", r.GetRoute().GetHost(), TransitionToIcon(r.GetTransition()),
				StateToIcon(getAggregatedStatus(r.GetStatus())))
		}
	}

	for _, i := range s.ConfigFiles {
		add(root, "ConfigFile", i.GetPath(), TransitionToIcon(i.GetTransition()),
			StateToIcon(getAggregatedStatus(i.GetStatus())))
	}

	for _, i := range s.CronJobs {
		add(root, "CronJob", i.GetJobName(), TransitionToIcon(i.GetTransition()), StateToIcon(i.GetLastExecution()))
	}

	if selectedNode != nil {
		tree.SetCurrentNode(selectedNode)
	} else {
		tree.SetCurrentNode(root)
	}

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			event = nil
		}
		return event
	})

	tree.Box.
		SetTitle("Platform Objects").
		SetBorder(true)

	return tree
}

func buildSelectedView(node *tview.TreeNode, status *capsule.Status) *tview.TextView {
	selectedView := tview.NewTextView().SetDynamicColors(true)

	selectedView.SetTitle(node.GetText()).
		SetBorder(true)

	text, state := getSelectedObjectTextAndState(node.GetText(), status)

	selectedView.SetText(tview.TranslateANSI(text))
	selectedView.SetTitleColor(StateToTCellColor(state))
	selectedView.SetBorderColor(StateToTCellColor(state))

	return selectedView
}

func getSelectedObjectTextAndState(name string, status *capsule.Status) (string, pipeline.ObjectState) {
	split := strings.SplitN(name, "/", 2)
	kind := split[0]
	split = strings.Split(split[1], " ")
	kindName := split[0]
	builder := &strings.Builder{}

	switch kind {
	case "Capsule":
		for _, s := range status.GetCapsule().GetStatuses() {
			GenericObjectStatusToString(s, builder)
		}
		return builder.String(), getAggregatedStatus(status.GetCapsule().GetStatuses())
	case "Interface":
		for _, i := range status.Interfaces {
			if i.GetName() == kindName {
				state := getAggregatedStatus(i.GetStatus())

				builder.WriteString(fmt.Sprintf("Port: %d\n", i.GetPort()))
				builder.WriteString(fmt.Sprintf("Transition: %s\n", TransitionToString(i.GetTransition())))
				builder.WriteString(fmt.Sprintf("State: %s\n", StateToString(state)))
				builder.WriteString("\nStatuses:\n")
				for _, s := range i.Status {
					GenericObjectStatusToString(s, builder)
				}
				return builder.String(), state
			}
		}
	case "Route":
		for _, i := range status.Interfaces {
			for _, r := range i.Routes {
				if r.GetRoute().GetHost() == kindName {
					state := getAggregatedStatus(r.GetStatus())

					builder.WriteString(fmt.Sprintf("ID: %s\n", r.GetRoute().GetId()))
					builder.WriteString(fmt.Sprintf("Transition: %s\n", TransitionToString(r.GetTransition())))
					builder.WriteString(fmt.Sprintf("State: %s\n", StateToString(state)))
					if len(r.GetRoute().GetPaths()) > 0 {
						builder.WriteString("Paths:\n")
						for _, p := range r.GetRoute().GetPaths() {
							builder.WriteString(GetIndented(fmt.Sprintf("%s %s", PathMatchToString(p.GetMatch()), p.GetPath()), 2))
						}
					}
					if len(r.GetRoute().GetOptions().GetAnnotations()) > 0 {
						builder.WriteString("Annotations:\n")
						for key, value := range r.GetRoute().GetOptions().GetAnnotations() {
							builder.WriteString(GetIndented(fmt.Sprintf("%s: %s", key, value), 2))
						}
					}

					builder.WriteString("\nStatuses:\n")
					for _, s := range r.Status {
						GenericObjectStatusToString(s, builder)
					}

					return builder.String(), state
				}
			}
		}
	case "ConfigFile":
		for _, c := range status.GetConfigFiles() {
			if c.GetPath() == kindName {
				state := getAggregatedStatus(c.GetStatus())

				builder.WriteString(fmt.Sprintf("Is secret: %t\n", c.GetIsSecret()))
				builder.WriteString(fmt.Sprintf("Transition: %s\n", TransitionToString(c.GetTransition())))
				builder.WriteString(fmt.Sprintf("State: %s\n", StateToString(state)))
				builder.WriteString("\nStatuses:\n")
				for _, s := range c.GetStatus() {
					GenericObjectStatusToString(s, builder)
				}
				return builder.String(), state
			}
		}
	case "CronJob":
		for _, c := range status.GetCronJobs() {
			if c.GetJobName() == kindName {
				builder.WriteString(fmt.Sprintf("Schedule: %s\n", c.GetSchedule()))
				builder.WriteString("Last execution:\n")
				builder.WriteString(fmt.Sprintf("Transition: %s\n", TransitionToString(c.GetTransition())))
				builder.WriteString(GetIndented(fmt.Sprintf("State: %s", StateToString(c.GetLastExecution())), 2))
				builder.WriteString("\nStatuses:\n")

				// TODO: INCLUDE STATUSES
				return builder.String(), pipeline.ObjectState_OBJECT_STATE_HEALTHY
			}

		}
	}

	return "", pipeline.ObjectState_OBJECT_STATE_UNSPECIFIED
}

// -------------------- Watching Functions --------------------

func watchCapsuleStatus(
	ctx context.Context,
	c *Cmd,
	capsuleID string,
	errChan chan error,
	statusChan chan *capsule.Status,
) {
	if !follow {
		statusResp, err := c.Rig.Capsule().GetStatus(ctx, connect.NewRequest(&capsule.GetStatusRequest{
			CapsuleId:     capsuleID,
			ProjectId:     flags.GetProject(c.Scope),
			EnvironmentId: flags.GetEnvironment(c.Scope),
		}))
		if err != nil {
			errChan <- err
			return
		}

		statusChan <- statusResp.Msg.GetStatus()
		return
	}

	stream, err := c.Rig.Capsule().WatchStatus(ctx, connect.NewRequest(&capsule.WatchStatusRequest{
		CapsuleId:     capsuleID,
		ProjectId:     flags.GetProject(c.Scope),
		EnvironmentId: flags.GetEnvironment(c.Scope),
	}))

	if err != nil {
		errChan <- err
		return
	}

	defer stream.Close()

	for stream.Receive() {
		select {
		case <-ctx.Done():
			return
		default:
			statusChan <- stream.Msg().GetStatus()
		}
	}

	if ctx.Err() != nil {
		return
	}

	if stream.Err() != nil {
		errChan <- stream.Err()
	}
}

func watchRollout(
	ctx context.Context,
	c *Cmd,
	capsuleID string,
	rolloutID uint64,
	errChan chan error,
	rolloutChan chan *capsule.Rollout,
) {
	if !follow {
		rolloutResp, err := c.Rig.Capsule().GetRollout(ctx, connect.NewRequest(&capsule.GetRolloutRequest{
			CapsuleId: capsuleID,
			RolloutId: rolloutID,
			ProjectId: flags.GetProject(c.Scope),
		}))
		if err != nil {
			errChan <- err
			return
		}

		rolloutChan <- rolloutResp.Msg.GetRollout()
		return
	}

	stream, err := c.Rig.Capsule().WatchRollouts(ctx, connect.NewRequest(&capsule.WatchRolloutsRequest{
		CapsuleId:     capsuleID,
		ProjectId:     flags.GetProject(c.Scope),
		EnvironmentId: flags.GetEnvironment(c.Scope),
		RolloutId:     rolloutID,
	}))
	if err != nil {
		errChan <- err
		return
	}

	defer stream.Close()

	for stream.Receive() {
		select {
		case <-ctx.Done():
			return
		default:
			rolloutChan <- stream.Msg().GetUpdated()
		}
	}

	if ctx.Err() != nil {
		return
	} else if stream.Err() != nil {
		errChan <- stream.Err()
	}
}
