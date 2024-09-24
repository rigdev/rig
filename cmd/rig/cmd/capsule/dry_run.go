package capsule

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"slices"

	"connectrpc.com/connect"
	"github.com/gdamore/tcell/v2"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rivo/tview"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
		return err
	}

	req.Changes = nil
	respOld, err := input.Rig.Capsule().Deploy(input.Ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}

	out, err := ProcessDryRunOutput(
		resp.Msg.GetOutcome(),
		respOld.Msg.GetOutcome(),
	)
	if err != nil {
		return err
	}

	if !input.IsInteractive {
		outputType := flags.Flags.OutputType
		if outputType == common.OutputTypePretty {
			outputType = common.OutputTypeYAML
		}
		return common.FormatPrint(out, outputType)
	}

	return PromptDryOutput(input.Ctx, out, input.Scheme)
}

type DryOutput struct {
	PlatformObjects   []PlatformDryObject
	KubernetesObjects []KubernetesDryObject
}

type PlatformDryObject struct {
	Old PlatformObject
	New PlatformObject
}

type PlatformObject struct {
	Object runtime.Object
	YAML   string
}

type KubernetesDryObject struct {
	Old KubernetesObject
	New KubernetesObject
}

type KubernetesObject struct {
	Object client.Object
	YAML   string
}

func ProcessDryRunOutput(
	newOutcome *capsule.DeployOutcome,
	oldOutcome *capsule.DeployOutcome,
) (DryOutput, error) {
	platforms := map[string]PlatformDryObject{}
	for _, o := range newOutcome.GetPlatformObjects() {
		ro, err := obj.DecodeUnstructuredRuntime([]byte(o.GetContentYaml()))
		if err != nil {
			return DryOutput{}, err
		}
		name := canonicalRuntimeObjName(ro)
		p := platforms[name]
		p.New = PlatformObject{
			Object: ro,
			YAML:   o.GetContentYaml(),
		}
		platforms[name] = p
	}
	for _, o := range oldOutcome.GetPlatformObjects() {
		ro, err := obj.DecodeUnstructuredRuntime([]byte(o.GetContentYaml()))
		if err != nil {
			return DryOutput{}, err
		}
		name := canonicalRuntimeObjName(ro)
		p := platforms[name]
		p.Old = PlatformObject{
			Object: ro,
			YAML:   o.GetContentYaml(),
		}
		platforms[name] = p
	}

	k8s := map[string]KubernetesDryObject{}
	for _, o := range newOutcome.GetKubernetesObjects() {
		co, err := obj.DecodeUnstructured([]byte(o.GetContentYaml()))
		if err != nil {
			return DryOutput{}, err
		}
		name := canonicalClientObjName(co)
		k := k8s[name]
		k.New = KubernetesObject{
			Object: co,
			YAML:   o.GetContentYaml(),
		}
		k8s[name] = k
	}
	for _, o := range oldOutcome.GetKubernetesObjects() {
		co, err := obj.DecodeUnstructured([]byte(o.GetContentYaml()))
		if err != nil {
			return DryOutput{}, err
		}
		name := canonicalClientObjName(co)
		k := k8s[name]
		k.Old = KubernetesObject{
			Object: co,
			YAML:   o.GetContentYaml(),
		}
		k8s[name] = k
	}

	result := DryOutput{}
	for _, key := range slices.Sorted(maps.Keys(platforms)) {
		result.PlatformObjects = append(result.PlatformObjects, platforms[key])
	}
	for _, key := range slices.Sorted(maps.Keys(k8s)) {
		result.KubernetesObjects = append(result.KubernetesObjects, k8s[key])
	}
	return result, nil
}

func canonicalRuntimeObjName(ro runtime.Object) string {
	return ro.GetObjectKind().GroupVersionKind().String()
}

func canonicalClientObjName(co client.Object) string {
	return fmt.Sprintf("%s %s %s", co.GetObjectKind().GroupVersionKind(), co.GetNamespace(), co.GetName())
}

type content struct {
	name string
	old  string
	new  string
	diff string
}

type state struct {
	itemIdx   int
	itemState int
	contents  []content
}

type view struct {
	list *tview.List
	text *tview.TextView
	mode *tview.TextView
	grid *tview.Grid
}

func makeTViewContent(outcome DryOutput, scheme *runtime.Scheme) ([]content, error) {
	var res []content
	for _, o := range outcome.PlatformObjects {
		var ro runtime.Object
		if o.New.Object != nil {
			ro = o.New.Object
		} else if o.Old.Object != nil {
			ro = o.Old.Object
		}
		name := fmt.Sprintf("[::b]%s", ro.GetObjectKind().GroupVersionKind().Kind)

		diff, err := makeTViewDiff(o.Old.Object, o.New.Object, scheme)
		if err != nil {
			return nil, err
		}

		oldYAML, err := common.ToYAMLColored(o.Old.YAML)
		if err != nil {
			return nil, err
		}

		newYAML, err := common.ToYAMLColored(o.New.YAML)
		if err != nil {
			return nil, err
		}

		res = append(res, content{
			name: name,
			old:  oldYAML,
			new:  newYAML,
			diff: diff,
		})
	}

	for _, o := range outcome.KubernetesObjects {
		name := fmt.Sprintf("%s/%s", o.New.Object.GetObjectKind().GroupVersionKind().Kind, o.New.Object.GetName())

		diff, err := makeTViewDiff(o.Old.Object, o.New.Object, scheme)
		if err != nil {
			return nil, err
		}

		oldYAML, err := common.ToYAMLColored(o.Old.YAML)
		if err != nil {
			return nil, err
		}

		newYAML, err := common.ToYAMLColored(o.New.YAML)
		if err != nil {
			return nil, err
		}

		res = append(res, content{
			name: name,
			old:  oldYAML,
			new:  newYAML,
			diff: diff,
		})
	}

	return res, nil
}

func PromptDryOutput(ctx context.Context, outcome DryOutput, scheme *runtime.Scheme) error {
	content, err := makeTViewContent(outcome, scheme)
	if err != nil {
		return err
	}

	s := state{
		itemIdx:   0,
		itemState: 0,
		contents:  content,
	}

	modeView := tview.NewTextView()
	modeView.SetTitle("Viewing Mode")
	modeView.SetBorder(true)
	modeView.SetDynamicColors(true)
	modeView.SetBackgroundColor(tcell.ColorNone)

	textView := tview.NewTextView()
	textView.SetTitle(fmt.Sprintf("%s (ESC or CTRL+C to cancel)", "Capsule Diff"))
	textView.SetBorder(true)
	textView.SetDynamicColors(true)
	textView.SetWrap(true)
	textView.SetBackgroundColor(tcell.ColorNone)

	listView := tview.NewList().ShowSecondaryText(false)
	listView.SetBorder(true).
		SetTitle("Resources (Return to view)")

	for _, c := range s.contents {
		listView.AddItem(c.name, "", 0, nil)
	}

	errChan := make(chan error)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grid := tview.NewGrid().SetRows(-1).SetColumns(-1, -2)
	grid.AddItem(modeView, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(textView, 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(listView, 0, 0, 2, 1, 0, 0, true)
	grid.SetRows(3, 0)

	view := view{
		list: listView,
		text: textView,
		mode: modeView,
		grid: grid,
	}
	updateView(s, view)

	app := tview.NewApplication().SetRoot(grid, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			s.itemIdx = listView.GetCurrentItem()
		case tcell.KeyRune:
			switch event.Rune() {
			case 'n':
				s.itemState = 0
			case 'c':
				s.itemState = 1
			case 'd':
				s.itemState = 2
			}
		case tcell.KeyEsc, tcell.KeyCtrlC:
			cancel()
		}
		updateView(s, view)
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

func makeTViewDiff(oldObj, newObj runtime.Object, scheme *runtime.Scheme) (string, error) {
	if oldObj == nil || newObj == nil {
		return "", nil
	}

	comparison := obj.NewComparison(oldObj, newObj, scheme)
	diff, err := comparison.ComputeDiff()
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	hr := &dyff.HumanReport{
		Report:     *diff.Report,
		OmitHeader: true,
	}
	if err := hr.WriteReport(tview.ANSIWriter(&out)); err != nil {
		return "", err
	}
	return out.String(), nil
}

func updateView(s state, view view) {
	// TODO There is probably a better tview compoennt to use than hacking in formatting text
	// based on state.
	var modeString string
	names := []string{"[N[]ew", "[C[]urrent", "[D[]iff"}
	for idx, n := range names {
		style := "[-:-:-:-]"
		if idx == s.itemState {
			style = "[::b]"
		}
		modeString += style + n + "  "
	}
	view.mode.SetText(modeString)

	c := s.contents[s.itemIdx]
	switch s.itemState {
	case 0:
		view.text.SetText(c.new)
	case 1:
		view.text.SetText(c.old)
	case 2:
		view.text.SetText(c.diff)
	}
}
