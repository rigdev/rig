package migrate

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rivo/tview"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReportSet struct {
	reports map[string]map[string]*dyff.Report
	scheme  *runtime.Scheme
}

func NewReportSet(scheme *runtime.Scheme) *ReportSet {
	return &ReportSet{
		reports: map[string]map[string]*dyff.Report{},
		scheme:  scheme,
	}
}

func (r *ReportSet) AddReport(original, proposal client.Object, reportName string) error {
	report, err := r.getDiffingReport(original, proposal)
	if err != nil {
		return err
	}

	if len(report.Diffs) == 0 {
		return nil
	}

	var kind string
	var name string
	if proposal != nil {
		kind = proposal.GetObjectKind().GroupVersionKind().Kind
		name = proposal.GetName()
	} else if original != nil {
		kind = original.GetObjectKind().GroupVersionKind().Kind
		name = original.GetName()
	}

	if reportName == "" {
		reportName = name
	}

	if _, ok := r.reports[kind]; !ok {
		r.reports[kind] = map[string]*dyff.Report{}
	}

	r.reports[kind][reportName] = report
	return nil
}

func (r *ReportSet) GetKind(kind string) (map[string]*dyff.Report, bool) {
	v, ok := r.reports[kind]
	return v, ok
}

func (r *ReportSet) GetKinds() []string {
	return maps.Keys(r.reports)
}

func (r *ReportSet) getDiffingReport(orig, proposal client.Object) (*dyff.Report, error) {
	c := obj.NewComparison(orig, proposal, r.scheme)
	c.AddFilter(obj.RemoveAnnotationsFilter(
		"kubectl.kubernetes.io/last-applied-configuration",
		"deployment.kubernetes.io/revision",
	))
	c.AddRemoveDiffs("status", "spec.template.spec.containers.*.env")
	d, err := c.ComputeDiff()
	if err != nil {
		return nil, err
	}

	return d.Report, nil
}

// marshall the platform resources into kubernetes resources, and then compare them to the existing k8s resources
func (c *Cmd) processPlatformOutput(
	migratedResources *Resources,
	outcome *capsule.DeployOutcome,
) error {
	for _, resource := range outcome.GetPlatformObjects() {
		proposal, err := obj.DecodeAny([]byte(resource.GetContentYaml()), c.Scheme)
		if err != nil {
			return err
		}

		if err := migratedResources.AddObject(proposal.GetObjectKind().GroupVersionKind().Kind,
			resource.GetName(),
			proposal); err != nil {
			return err
		}
	}

	for _, out := range outcome.GetKubernetesObjects() {
		proposal, err := obj.DecodeAny([]byte(out.GetContentYaml()), c.Scheme)
		if err != nil {
			return fmt.Errorf("error decoding object from operator: %v", err)
		}

		if err := migratedResources.AddObject(proposal.GetObjectKind().GroupVersionKind().Kind,
			out.GetName(),
			proposal); err != nil {
			return fmt.Errorf("error adding 'migrated' object': %v", err)
		}
	}

	return nil
}

func ProcessOperatorOutput(
	migratedResources *Resources,
	operatorOutput []*pipeline.ObjectChange,
	scheme *runtime.Scheme,
) error {
	for _, out := range operatorOutput {
		proposal, err := obj.DecodeAny([]byte(out.GetObject().GetContent()), scheme)
		if err != nil {
			return fmt.Errorf("error decoding object from operator: %v", err)
		}

		if err := migratedResources.AddObject(proposal.GetObjectKind().GroupVersionKind().Kind,
			proposal.GetName(),
			proposal); err != nil {
			return fmt.Errorf("error adding 'migrated' object': %v", err)
		}
	}

	return nil
}

func ProcessHelmOutput(
	helmOutput map[string]string,
	scheme *runtime.Scheme,
) ([]client.Object, error) {
	var objects []client.Object
	for _, yml := range helmOutput {
		decoder := yaml.NewDecoder(bytes.NewBufferString(yml))
		for {
			out := map[string]any{}
			err := decoder.Decode(&out)
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			bs, err := yaml.Marshal(out)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(string(bs)) == "" {
				continue
			}
			proposal, err := obj.DecodeAny(bs, scheme)
			if err != nil {
				return nil, fmt.Errorf("error decoding object from helm: %v", err)
			}

			objects = append(objects, proposal)
		}

	}

	return objects, nil
}

func getWarningsView(warnings []*Warning) *tview.TextView {
	if len(warnings) == 0 {
		return nil
	}
	var out bytes.Buffer
	for _, w := range warnings {
		out.WriteString(w.String())
		out.WriteString("\n")
	}

	text := tview.NewTextView()
	text.SetTitle("Warnings (ESC to remove)")
	text.SetBorder(true)
	text.SetDynamicColors(true)
	text.SetWrap(true)
	text.SetTextColor(tcell.ColorYellow)
	text.SetText(out.String())
	text.SetBackgroundColor(tcell.ColorNone)

	return text
}

func showOverview(
	currentOverview *tview.TreeView,
	migratedOverview *tview.TreeView,
) error {
	currentOverview.Box.SetTitleColor(tcell.ColorRed).SetBorderColor(tcell.ColorRed)
	migratedOverview.Box.SetTitleColor(tcell.ColorGreen).SetBorderColor(tcell.ColorGreen)

	grid := tview.NewGrid().
		SetRows(0).
		SetColumns(0, 0).
		AddItem(currentOverview, 0, 0, 10, 1, 0, 0, false).
		AddItem(migratedOverview, 0, 1, 10, 1, 0, 0, false)
	app := tview.NewApplication().SetRoot(grid, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.Stop()
		}
		return event
	})

	return app.Run()
}

func showCapsule(capsule *platformv1.Capsule) error {
	capsuleYaml, err := common.Format(capsule, common.OutputTypeYAML)
	if err != nil {
		return err
	}

	text := tview.NewTextView()
	text.SetTitle("Platform Capsule (ESC to exit)")
	text.SetBorder(true)
	text.SetDynamicColors(true)
	text.SetWrap(true)
	text.SetText(capsuleYaml)
	text.SetBackgroundColor(tcell.ColorNone)

	app := tview.NewApplication().SetRoot(text, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.Stop()
		}
		return event
	})

	return app.Run()
}

// func showRaw(yaml, kind, name string) error {
// 	text := tview.NewTextView()
// 	text.SetTitle(fmt.Sprintf("%s/%s (ESC to exit)", kind, name))
// 	text.SetBorder(true)
// 	text.SetDynamicColors(true)
// 	text.SetWrap(true)
// 	text.SetText(yaml)
// 	text.SetBackgroundColor(tcell.ColorNone)

// 	app := tview.NewApplication().SetRoot(text, true)
// 	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
// 		if event.Key() == tcell.KeyESC {
// 			app.Stop()
// 		}
// 		return event
// 	})

// 	return app.Run()
// }

func ShowDiffReport(r *dyff.Report, kind, name string, warnings []*Warning) error {
	var text *tview.TextView
	if r != nil {
		var out bytes.Buffer
		hr := &dyff.HumanReport{
			Report:     *r,
			OmitHeader: true,
		}
		if err := hr.WriteReport(tview.ANSIWriter(&out)); err != nil {
			return err
		}

		text = tview.NewTextView()
		text.SetTitle(fmt.Sprintf("%s/%s (ESC to exit)", kind, name))
		text.SetBorder(true)
		text.SetDynamicColors(true)
		text.SetWrap(true)
		text.SetText(out.String())
		text.SetBackgroundColor(tcell.ColorNone)
	}

	warningsText := getWarningsView(warnings)

	grid := tview.NewGrid().
		SetColumns(0).
		SetBorders(false)

	if warningsText != nil && text != nil {
		grid.SetRows(-1, -2).
			AddItem(warningsText, 0, 0, 1, 1, 0, 0, false).
			AddItem(text, 1, 0, 1, 1, 0, 0, true)
	} else if text != nil {
		grid.SetRows(0).
			AddItem(text, 0, 0, 1, 1, 0, 0, true)
	} else {
		grid.SetRows(0).
			AddItem(warningsText, 0, 0, 1, 1, 0, 0, false)
	}

	app := tview.NewApplication().SetRoot(grid, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			prim := app.GetFocus()
			if prim == text || text == nil {
				app.Stop()
			} else {
				grid.RemoveItem(warningsText)
				grid.RemoveItem(text)
				grid.AddItem(text, 0, 0, 2, 1, 0, 0, true)
				app.SetFocus(text)
			}

		}
		return event
	})

	return app.Run()
}

func CreateMigratedOverview(reportSet *ReportSet) *tview.TreeView {
	capsuleRoot := "Capsule/"
	for name := range reportSet.reports["Capsule"] {
		capsuleRoot = capsuleRoot + name
	}

	root := tview.NewTreeNode(capsuleRoot).SetSelectable(false)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	add := func(parent *tview.TreeNode, kind string, name string) *tview.TreeNode {
		node := tview.NewTreeNode(fmt.Sprintf("%s/%s", kind, name)).
			SetSelectable(false)

		parent.AddChild(node)
		return node
	}

	deploymentName := ""
	for name := range reportSet.reports["Deployment"] {
		deploymentName = name
	}

	deploymentNode := add(root, "Deployment", deploymentName)

	serviceName := ""
	for name := range reportSet.reports["Service"] {
		serviceName = name
	}

	var serviceNode *tview.TreeNode
	if serviceName != "" {
		serviceNode = add(deploymentNode, "Service", serviceName)
	}

	for kind, reports := range reportSet.reports {
		if kind == "Deployment" || kind == "Capsule" || kind == "Service" {
			continue
		}

		names := make([]string, 0, len(reports))
		for name := range reports {
			names = append(names, name)
		}

		switch kind {
		case "ServiceAccount":
			for _, name := range names {
				add(deploymentNode, kind, name)
			}
		case "ConfigMap":
			for _, name := range names {
				add(deploymentNode, kind, name)
			}
		case "Secret":
			for _, name := range names {
				add(deploymentNode, kind, name)
			}
		case "Ingress":
			for _, name := range names {
				add(serviceNode, kind, name)
			}
		case "CronJob":
			for _, name := range names {
				add(root, kind, name)
			}
		case "HorizontalPodAutoscaler":
			for _, name := range names {
				add(deploymentNode, kind, name)
			}
		}
	}

	tree.Box.SetTitle("Migrated Resources (ESC to exit)").
		SetTitleColor(tcell.ColorGreen).
		SetBorder(true).
		SetBorderColor(tcell.ColorGreen)

	return tree
}
