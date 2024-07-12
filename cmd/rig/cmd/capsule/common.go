package capsule

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	api_rollout "github.com/rigdev/rig-go-api/api/v1/capsule/rollout"
	"github.com/rigdev/rig-go-api/model"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var CapsuleID string

const (
	BasicGroupID           = "basic"
	DeploymentGroupID      = "deployment"
	TroubleshootingGroupID = "troubleshooting"
)

// deployment structs
type BaseInput struct {
	Ctx           context.Context
	Rig           rig.Client
	ProjectID     string
	EnvironmentID string
	CapsuleID     string
}

type DeployInput struct {
	BaseInput
	Changes            []*capsule.Change
	ForceDeploy        bool
	ForceOverride      bool
	CurrentRolloutID   uint64
	CurrentFingerprint *model.Fingerprint
}

type DeployAndWaitInput struct {
	DeployInput
	Timeout    time.Duration
	RollbackID uint64
	NoWait     bool
}

type DeployDryInput struct {
	BaseInput
	Changes            []*capsule.Change
	CurrentRolloutID   uint64
	CurrentFingerprint *model.Fingerprint
	Scheme             *runtime.Scheme
	IsInteractive      bool
}

type RollbackInput struct {
	BaseInput
	CurrentRolloutID uint64
	RollbackID       uint64
}

type WaitForRolloutInput struct {
	RollbackInput
	Revision *capsule.Revision
	Timeout  time.Duration
}

type GetRolloutInput struct {
	BaseInput
	RolloutID uint64
}

func GetCurrentContainerResources(
	ctx context.Context,
	client rig.Client,
	scope scope.Scope,
) (*capsule.ContainerSettings, uint32, error) {
	rollout, err := GetCurrentRollout(ctx, client, scope)
	if err != nil {
		return nil, 0, err
	}
	container := rollout.GetConfig().GetContainerSettings()
	if container == nil {
		container = &capsule.ContainerSettings{}
	}
	if container.Resources == nil {
		container.Resources = &capsule.Resources{}
	}

	utils.FeedDefaultResources(container.Resources)

	return container, rollout.GetConfig().GetReplicas(), nil
}

func GetCurrentNetwork(ctx context.Context, client rig.Client, scope scope.Scope) (*capsule.Network, error) {
	rollout, err := GetCurrentRollout(ctx, client, scope)
	if err != nil {
		return nil, err
	}
	return rollout.GetConfig().GetNetwork(), nil
}

func GetCurrentRollout(ctx context.Context, client rig.Client, scope scope.Scope) (*capsule.Rollout, error) {
	return GetCurrentRolloutOfCapsule(ctx, client, scope, CapsuleID)
}

func GetCurrentRolloutOfCapsule(
	ctx context.Context,
	client rig.Client,
	scope scope.Scope,
	capsuleID string,
) (*capsule.Rollout, error) {
	r, err := client.Capsule().ListRollouts(ctx, connect.NewRequest(&capsule.ListRolloutsRequest{
		CapsuleId: capsuleID,
		Pagination: &model.Pagination{
			Offset:     0,
			Limit:      1,
			Descending: true,
		},
		ProjectId:     flags.GetProject(scope),
		EnvironmentId: flags.GetEnvironment(scope),
	}))
	if err != nil {
		return nil, err
	}

	for _, r := range r.Msg.GetRollouts() {
		return r, nil
	}

	return nil, errors.NotFoundErrorf("no rollout for capsule")
}

func Truncated(str string, max int) string {
	if len(str) > max {
		return str[:strings.LastIndexAny(str[:max], " .,:;-")] + "..."
	}

	return str
}

func TruncatedFixed(str string, max int) string {
	if len(str) > max {
		return str[:max] + "..."
	}

	return str
}

func PromptAbortAndDeploy(
	ctx context.Context,
	rig rig.Client,
	prompter common.Prompter,
	req *connect.Request[capsule.DeployRequest],
) (*connect.Response[capsule.DeployResponse], error) {
	deploy, err := prompter.Confirm("Rollout already in progress, would you like to cancel it and redeploy?", false)
	if err != nil {
		return nil, err
	}

	if !deploy {
		return nil, errors.FailedPreconditionErrorf("rollout already in progress")
	}

	return AbortAndDeploy(ctx, rig, req)
}

func AbortAndDeploy(
	ctx context.Context,
	rig rig.Client,
	req *connect.Request[capsule.DeployRequest],
) (*connect.Response[capsule.DeployResponse], error) {
	req.Msg.Force = true
	return rig.Capsule().Deploy(ctx, req)
}

func PrintLogs(stream *connect.ServerStreamForClient[capsule.LogsResponse]) error {
	for stream.Receive() {
		switch v := stream.Msg().GetLog().GetMessage().GetMessage().(type) {
		case *capsule.LogMessage_Stdout:
			if err := printInstanceID(stream.Msg().GetLog().GetInstanceId(), os.Stdout); err != nil {
				return err
			}
			os.Stdout.WriteString(stream.Msg().GetLog().GetTimestamp().AsTime().Format(cli.RFC3339NanoFixed))
			os.Stdout.WriteString(": ")
			if _, err := os.Stdout.Write(v.Stdout); err != nil {
				return err
			}
		case *capsule.LogMessage_Stderr:
			if err := printInstanceID(stream.Msg().GetLog().GetInstanceId(), os.Stderr); err != nil {
				return err
			}
			os.Stderr.WriteString(stream.Msg().GetLog().GetTimestamp().AsTime().Format(cli.RFC3339NanoFixed))
			os.Stderr.WriteString(": ")
			if _, err := os.Stderr.Write(v.Stderr); err != nil {
				return err
			}
		case *capsule.LogMessage_ContainerTermination_:
			if err := printInstanceID(stream.Msg().GetLog().GetInstanceId(), os.Stderr); err != nil {
				return err
			}
			os.Stdout.WriteString(stream.Msg().GetLog().GetTimestamp().AsTime().Format(cli.RFC3339NanoFixed))
			os.Stdout.WriteString(" Container Terminated.\n\n")
		default:
			return errors.InvalidArgumentErrorf("invalid log message")
		}
	}

	return stream.Err()
}

func SelectCapsule(ctx context.Context, rc rig.Client, prompter common.Prompter, scope scope.Scope) (string, error) {
	resp, err := rc.Capsule().List(ctx, connect.NewRequest(&capsule.ListRequest{
		Pagination: &model.Pagination{},
		ProjectId:  flags.GetProject(scope),
	}))
	if err != nil {
		return "", err
	}

	var capsuleNames []string
	for _, c := range resp.Msg.GetCapsules() {
		capsuleNames = append(capsuleNames, c.GetCapsuleId())
	}

	if len(capsuleNames) == 0 {
		return "", errors.New("This project has no capsules. Create one, to get started")
	}

	_, name, err := prompter.Select("Capsule: ", capsuleNames, common.SelectFuzzyFilterOpt)
	if err != nil {
		return "", err
	}

	return name, nil
}

var colors = []color.Attribute{
	color.FgRed,
	color.FgBlue,
	color.FgCyan,
	color.FgGreen,
	color.FgYellow,
	color.FgMagenta,
	color.FgWhite,
}

var instanceToColor = map[string]color.Attribute{}

func printInstanceID(instanceID string, out *os.File) error {
	c, ok := instanceToColor[instanceID]
	if !ok {
		c = colors[len(instanceToColor)%len(colors)]
		instanceToColor[instanceID] = c
	}
	color.Set(c)
	if _, err := out.WriteString(instanceID + " "); err != nil {
		return fmt.Errorf("could not print instance id: %w", err)
	}
	color.Unset()
	return nil
}

func Deploy(input DeployInput) (*capsule.Revision, uint64, error) {
	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId:          input.CapsuleID,
			Changes:            input.Changes,
			Force:              input.ForceDeploy,
			ProjectId:          input.ProjectID,
			EnvironmentId:      input.EnvironmentID,
			DryRun:             false,
			CurrentRolloutId:   input.CurrentRolloutID,
			ForceOverride:      input.ForceOverride,
			CurrentFingerprint: input.CurrentFingerprint,
		},
	}

	res, err := input.Rig.Capsule().Deploy(input.Ctx, req)
	if err != nil {
		return nil, 0, err
	}

	return res.Msg.GetRevision(), res.Msg.GetRolloutId(), nil
}

func DeployAndWait(
	input DeployAndWaitInput,
) error {
	revision, rolloutID, err := Deploy(input.DeployInput)
	if err != nil {
		return err
	}
	fmt.Println("Deploying to capsule", input.CapsuleID)
	if input.NoWait {
		return nil
	}

	waitForRolloutInput := WaitForRolloutInput{
		RollbackInput: RollbackInput{
			BaseInput:        input.BaseInput,
			CurrentRolloutID: rolloutID,
			RollbackID:       input.RollbackID,
		},
		Revision: revision,
		Timeout:  input.Timeout,
	}
	return WaitForRollout(waitForRolloutInput)
}

func DeployDry(input DeployDryInput) error {
	req := &capsule.DeployRequest{
		CapsuleId:          input.CapsuleID,
		Changes:            input.Changes,
		ProjectId:          input.ProjectID,
		EnvironmentId:      input.EnvironmentID,
		DryRun:             true,
		CurrentRolloutId:   input.CurrentRolloutID,
		CurrentFingerprint: input.CurrentFingerprint,
	}

	resp, err := input.Rig.Capsule().Deploy(input.Ctx, connect.NewRequest(req))
	if err != nil {
		return fmt.Errorf("failed to get new stuff: %w", err)
	}
	outcome := resp.Msg.GetOutcome()

	var out DryOutput
	out.PlatformCapsule = resp.Msg.GetRevision().GetSpec()

	for _, o := range outcome.GetPlatformObjects() {
		co, err := obj.DecodeAny([]byte(o.GetContentYaml()), input.Scheme)
		if err != nil {
			return err
		}
		out.DirectKubernetes = append(out.DirectKubernetes, co)
	}
	for _, o := range outcome.GetKubernetesObjects() {
		co, err := obj.DecodeAny([]byte(o.GetContentYaml()), input.Scheme)
		if err != nil {
			return err
		}
		out.DerivedKubernetes = append(out.DerivedKubernetes, co)
	}

	if !input.IsInteractive {
		outputType := flags.Flags.OutputType
		if outputType == common.OutputTypePretty {
			outputType = common.OutputTypeYAML
		}
		return common.FormatPrint(out, outputType)
	}

	return promptDryOutput(input.Ctx, out, outcome)
}

func promptDryOutput(ctx context.Context, out DryOutput, outcome *capsule.DeployOutcome) error {
	listView := tview.NewList().ShowSecondaryText(false)
	listView.SetBorder(true).
		SetTitle("Resources (Return to view)")

	capsuleYamlBytes, err := yaml.Marshal(out.PlatformCapsule)
	if err != nil {
		return err
	}

	capsuleYaml := string(capsuleYamlBytes)

	listView.AddItem("[::b]Platform Capsule", "", 0, nil)
	content := []string{capsuleYaml}
	for i, o := range out.DirectKubernetes {
		obj := o.(client.Object)
		listView.AddItem(fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName()), "", 0, nil)
		content = append(content, outcome.PlatformObjects[i].ContentYaml)
	}
	for i, o := range out.DerivedKubernetes {
		obj := o.(client.Object)
		listView.AddItem(fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName()), "", 0, nil)
		content = append(content, outcome.KubernetesObjects[i].ContentYaml)
	}

	textView := tview.NewTextView()
	textView.SetTitle(fmt.Sprintf("%s (ESC to continue and CTRL+C to cancel)", "Capsule Diff"))
	textView.SetBorder(true)
	textView.SetDynamicColors(true)
	textView.SetWrap(true)
	textView.SetText(capsuleYaml)
	textView.SetBackgroundColor(tcell.ColorNone)

	errChan := make(chan error)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grid := tview.NewGrid().SetRows(-1).SetColumns(-1, -2)
	grid.AddItem(textView, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(listView, 0, 0, 1, 1, 0, 0, true)

	app := tview.NewApplication().SetRoot(grid, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			textView.SetText(content[listView.GetCurrentItem()])
		case tcell.KeyEsc:
			cancel()
		case tcell.KeyCtrlC:
			cancel()
		}
		return event
	})

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
		case <-ctx.Done():
			return nil
		}
	}
}

type DryOutput struct {
	PlatformCapsule   *platformv1.Capsule `json:"platformCapsule"`
	DirectKubernetes  []any               `json:"directKubernetes"`
	DerivedKubernetes []any               `json:"derivedKubernetes"`
}

func Rollback(input RollbackInput) (*capsule.Revision, uint64, error) {
	deployInput := DeployInput{
		BaseInput: input.BaseInput,
		Changes: []*capsule.Change{
			{
				Field: &capsule.Change_Rollback_{
					Rollback: &capsule.Change_Rollback{
						RollbackId: input.RollbackID,
					},
				},
			},
		},
		ForceDeploy:      true,
		CurrentRolloutID: input.CurrentRolloutID,
	}

	return Deploy(deployInput)
}

func WaitForRollout(
	input WaitForRolloutInput,
) error {
	if input.CurrentRolloutID == 0 {
		first := true
		for {
			resp, err := input.Rig.Capsule().GetRolloutOfRevisions(
				input.Ctx,
				connect.NewRequest(&capsule.GetRolloutOfRevisionsRequest{
					ProjectId:     input.ProjectID,
					EnvironmentId: input.EnvironmentID,
					CapsuleId:     input.CapsuleID,
					Fingerprints: &model.Fingerprints{
						Capsule: input.Revision.GetMetadata().GetFingerprint(),
					},
				}))
			if err != nil {
				return err
			}
			switch r := resp.Msg.GetKind().(type) {
			case *capsule.GetRolloutOfRevisionsResponse_NoRollout_:
			case *capsule.GetRolloutOfRevisionsResponse_Rollout:
				input.CurrentRolloutID = r.Rollout.GetRolloutId()
			}
			if input.CurrentRolloutID == 0 {
				if first {
					fmt.Println("Waiting for rollout to start...")
					first = false
				}
			} else {
				break
			}
			time.Sleep(time.Second)
		}
	}

	fmt.Printf("Rollout %v started\n", input.CurrentRolloutID)

	var deadline time.Time
	if input.Timeout != 0 {
		deadline = time.Now().Add(input.Timeout)
	}

	var lastConfigure []*api_rollout.StepInfo
	var lastResource []*api_rollout.StepInfo
	var lastRunning []*api_rollout.StepInfo
	for {
		if !deadline.IsZero() && time.Now().After(deadline) {
			fmt.Println()
			fmt.Printf("🛑 Rollout timed out after %s... ", input.Timeout)
			if input.RollbackID == 0 {
				return fmt.Errorf("aborted")
			}

			_, _, err := Rollback(input.RollbackInput)
			if err != nil {
				return err
			}

			return fmt.Errorf("rollback to %d initiated", input.RollbackID)
		}

		getRolloutInput := GetRolloutInput{
			BaseInput: input.BaseInput,
			RolloutID: input.CurrentRolloutID,
		}

		rollout, err := getRollout(getRolloutInput)
		if err != nil {
			return err
		}

		// Check if the rollout was stopped by the user or if another rollout was started in the meantime.
		if rollout.GetStatus().GetState() == api_rollout.State_STATE_STOPPED {
			str := "🛑 Rollout"
			switch rollout.GetStatus().GetResult() {
			case api_rollout.Result_RESULT_REPLACED:
				str += " was replaced by a later rollout"
			case api_rollout.Result_RESULT_ABORTED:
				str += " was aborted"
			default:
				str += " was stopped"
			}

			fmt.Println(str)
			os.Exit(1)

			return nil
		}

		var configure []*api_rollout.StepInfo
		if stage := rollout.GetStatus().GetStages().GetConfigure(); stage != nil {
			for _, s := range stage.GetSteps() {
				var info *api_rollout.StepInfo

				switch v := s.GetStep().(type) {
				case *api_rollout.ConfigureStep_Generic:
					info = v.Generic.GetInfo()
				case *api_rollout.ConfigureStep_Commit:
					info = v.Commit.GetInfo()
				case *api_rollout.ConfigureStep_ConfigureCapsule:
					info = v.ConfigureCapsule.GetInfo()
				case *api_rollout.ConfigureStep_ConfigureEnv:
					info = v.ConfigureEnv.GetInfo()
				case *api_rollout.ConfigureStep_ConfigureFile:
					info = v.ConfigureFile.GetInfo()
				}

				if info != nil {
					configure = append(configure, info)
				}
			}
		}
		var resource []*api_rollout.StepInfo
		if stage := rollout.GetStatus().GetStages().GetResourceCreation(); stage != nil {
			for _, s := range stage.GetSteps() {
				var info *api_rollout.StepInfo

				switch v := s.GetStep().(type) {
				case *api_rollout.ResourceCreationStep_Generic:
					info = v.Generic.GetInfo()
				case *api_rollout.ResourceCreationStep_CreateResource:
					info = v.CreateResource.GetInfo()
				}

				if info != nil {
					resource = append(resource, info)
				}
			}
		}

		var running []*api_rollout.StepInfo
		done := false
		if stage := rollout.GetStatus().GetStages().GetRunning(); stage != nil {
			done = true
			for _, s := range stage.GetSteps() {
				var info *api_rollout.StepInfo

				switch v := s.GetStep().(type) {
				case *api_rollout.RunningStep_Generic:
					info = v.Generic.GetInfo()

					if info.GetState() != api_rollout.StepState_STEP_STATE_DONE {
						done = false
					}

				case *api_rollout.RunningStep_Instances:
					info = v.Instances.GetInfo()

					if info.GetState() != api_rollout.StepState_STEP_STATE_DONE {
						done = false
					}
				}

				if info != nil {
					running = append(running, info)
				}
			}
		}

		printSteps := func(steps []*api_rollout.StepInfo) {
			for _, s := range steps {
				icon := "❔"
				msg := ""
				switch s.GetState() {
				case api_rollout.StepState_STEP_STATE_FAILED:
					icon = "🚫"
					msg = "Failed"
				case api_rollout.StepState_STEP_STATE_ONGOING:
					icon = "⏳"
					msg = "Ongoing"
				case api_rollout.StepState_STEP_STATE_DONE:
					icon = "✅"
					msg = "Done"
				}

				if s.GetMessage() != "" {
					msg = s.GetMessage()
				}

				fmt.Printf("%s %s: %s\n", icon, s.GetName(), msg)
			}
		}

		printNewSteps := func(current, last []*api_rollout.StepInfo) []*api_rollout.StepInfo {
			var newSteps []*api_rollout.StepInfo
			for _, s := range current {
				found := false
				for _, l := range last {
					if l.GetName() == s.GetName() {
						if l.GetState() == s.GetState() && l.GetMessage() == s.GetMessage() {
							found = true
							break
						}
					}
				}
				if !found {
					newSteps = append(newSteps, s)
				}
			}
			printSteps(newSteps)
			return newSteps
		}

		lastConfigure = append(lastConfigure, printNewSteps(configure, lastConfigure)...)
		lastResource = append(lastResource, printNewSteps(resource, lastResource)...)
		lastRunning = append(lastRunning, printNewSteps(running, lastRunning)...)

		if done {
			fmt.Println("")
			fmt.Println("Done ✅ - Rollout Complete")
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}

func getRollout(
	input GetRolloutInput,
) (*capsule.Rollout, error) {
	connectionLost := false
	ctx, cancel := context.WithDeadline(input.Ctx, time.Now().Add(2*time.Minute))
	defer cancel()

	for {
		res, err := input.Rig.Capsule().GetRollout(ctx, &connect.Request[capsule.GetRolloutRequest]{
			Msg: &capsule.GetRolloutRequest{
				CapsuleId: input.CapsuleID,
				RolloutId: input.RolloutID,
				ProjectId: input.ProjectID,
			},
		})
		if errors.IsUnavailable(err) || ctx.Err() != nil {
			if !connectionLost {
				fmt.Println("🚫 Connection lost, retrying...")
				connectionLost = true
				time.Sleep(1 * time.Second)
				continue
			}

			// Check if deadling exceeded
			if ctx.Err() != nil {
				return nil, errors.UnavailableErrorf("🚫 Failed to restore the connection")
			}
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			return nil, err
		}

		if connectionLost {
			fmt.Println("✅ Connection restored")
		}

		return res.Msg.GetRollout(), nil
	}
}

func PortForward(
	ctx context.Context,
	rig rig.Client,
	scope scope.Scope,
	capsuleID, instanceID string,
	localPort uint32,
	remotePort uint32,
	verbose bool,
) error {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return err
	}

	return PortForwardOnListener(ctx, rig, scope, capsuleID, instanceID, l, remotePort, verbose)
}

func PortForwardOnListener(
	ctx context.Context,
	rig rig.Client,
	scope scope.Scope,
	capsuleID, instanceID string,
	l net.Listener,
	remotePort uint32,
	verbose bool,
) error {
	fmt.Printf("[rig] connected to instance '%s', accepting traffic on %s\n", instanceID, l.Addr().String())

	for {
		conn, err := l.Accept()
		if err != nil {
			if verbose {
				return fmt.Errorf("error listening for incoming connections: %v", err)
			}
		}

		if verbose {
			fmt.Printf("[rig] new connection %s -> %s:%d\n", conn.RemoteAddr().String(), instanceID, remotePort)
		}

		go func() {
			err := runPortForwardForPort(ctx, rig, scope, capsuleID, instanceID, conn, remotePort, verbose)
			if errors.IsNotFound(err) {
				fmt.Printf("[rig] instance '%s' no longer available: %v\n", instanceID, err)
				os.Exit(1)
			} else if err != nil {
				fmt.Println("[rig] connection closed with error:", err)
			} else if verbose {
				fmt.Printf("[rig] closed connection %s -> %s:%d\n", conn.RemoteAddr().String(), instanceID, remotePort)
			}
		}()
	}
}

func runPortForwardForPort(
	ctx context.Context,
	rig rig.Client,
	scope scope.Scope,
	capsuleID, instanceID string,
	conn net.Conn,
	port uint32,
	verbose bool,
) error {
	defer conn.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pf := rig.Capsule().PortForward(ctx)
	if err := pf.Send(&capsule.PortForwardRequest{
		Request: &capsule.PortForwardRequest_Start_{
			Start: &capsule.PortForwardRequest_Start{
				ProjectId:     flags.GetProject(scope),
				EnvironmentId: flags.GetEnvironment(scope),
				CapsuleId:     capsuleID,
				InstanceId:    instanceID,
				Port:          port,
			},
		},
	}); err != nil {
		return err
	}

	go func() {
		for {
			buff := make([]byte, 32*1024)
			n, err := conn.Read(buff)
			if err == io.EOF {
				if err := pf.Send(&capsule.PortForwardRequest{Request: &capsule.PortForwardRequest_Close_{
					Close: &capsule.PortForwardRequest_Close{},
				}}); err != nil {
					cancel()
					if verbose {
						fmt.Println("[rig] error sending close:", err)
					}
					return
				}
				return
			} else if err != nil {
				cancel()
				return
			}

			if err := pf.Send(&capsule.PortForwardRequest{Request: &capsule.PortForwardRequest_Data{
				Data: buff[:n],
			}}); err != nil {
				cancel()
				if verbose {
					fmt.Println("[rig] error sending data to server:", err)
				}
				return
			}
		}
	}()

	for {
		res, err := pf.Receive()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		switch v := res.GetResponse().(type) {
		case *capsule.PortForwardResponse_Data:
			if _, err := conn.Write(v.Data); err != nil {
				return err
			}
		case *capsule.PortForwardResponse_Close_:
			return nil
		}
	}
}

func GetCapsuleInstance(
	ctx context.Context,
	rig rig.Client,
	scope scope.Scope,
	capsuleID string,
) (string, error) {
	instancesRes, err := rig.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
		Msg: &capsule.ListInstancesRequest{
			ProjectId:     flags.GetProject(scope),
			EnvironmentId: flags.GetEnvironment(scope),
			CapsuleId:     capsuleID,
			Pagination: &model.Pagination{
				Limit: 1,
			},
		},
	})
	if err != nil {
		return "", err
	}

	if len(instancesRes.Msg.Instances) == 0 {
		return "", errors.NotFoundErrorf("no instances found for capsule")
	}

	return instancesRes.Msg.Instances[0].GetInstanceId(), nil
}
