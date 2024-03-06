package migrate

import (
	"bytes"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rivo/tview"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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

func (r *ReportSet) AddObject(original, proposal client.Object) error {
	report, err := r.getDiffingReport(original, proposal)
	if err != nil {
		return err
	}

	kind := original.GetObjectKind().GroupVersionKind().Kind
	if _, ok := r.reports[kind]; !ok {
		r.reports[kind] = map[string]*dyff.Report{}
	}

	r.reports[kind][original.GetName()] = report
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
func processPlatformOutput(
	reports *ReportSet,
	currentResources *CurrentResources,
	platformResources map[string]string,
	scheme *runtime.Scheme,
) (*v1alpha2.Capsule, error) {
	var capsule *v1alpha2.Capsule
	for _, resource := range platformResources {
		// unmarshal the resource into a k8s object
		object := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(resource), object); err != nil {
			return nil, err
		}

		orig := currentResources.getCurrentObject(object.GetKind(), object.GetName())

		proposal, err := obj.DecodeAny([]byte(resource), scheme)
		if err != nil {
			return nil, err
		}

		if err := reports.AddObject(orig, proposal); err != nil {
			return nil, err
		}

		if object.GetKind() == "Capsule" {
			capsule = &v1alpha2.Capsule{}
			if err = obj.Decode([]byte(resource), capsule); err != nil {
				return nil, err
			}
		}
	}

	return capsule, nil
}

func ProcessOperatorOutput(
	reports *ReportSet,
	currentResources *CurrentResources,
	operatorOutput []*pipeline.ObjectChange,
	scheme *runtime.Scheme,
) error {
	for _, out := range operatorOutput {
		orig := currentResources.getCurrentObject(out.GetObject().GetGvk().GetKind(), out.GetObject().GetName())

		proposal, err := obj.DecodeAny([]byte(out.GetObject().GetContent()), scheme)
		if err != nil {
			return err
		}
		if err := reports.AddObject(orig, proposal); err != nil {
			return err
		}
	}

	return nil
}

func processRemainingResources(
	reports *ReportSet,
	currentResources *CurrentResources,
	scheme *runtime.Scheme,
) error {
	if currentResources.Deployment != nil {
		if err := reports.AddObject(currentResources.Deployment, nil); err != nil {
			return err
		}
	}

	if currentResources.HPA != nil {
		if err := reports.AddObject(currentResources.HPA, nil); err != nil {
			return err
		}
	}

	if currentResources.ServiceAccount != nil {
		if err := reports.AddObject(currentResources.ServiceAccount, nil); err != nil {
			return err
		}
	}

	if currentResources.Capsule != nil {
		if err := reports.AddObject(currentResources.Capsule, nil); err != nil {
			return err
		}
	}

	for _, cm := range currentResources.ConfigMaps {
		if err := reports.AddObject(cm, nil); err != nil {
			return err
		}
	}

	for _, s := range currentResources.Secrets {
		if err := reports.AddObject(s, nil); err != nil {
			return err
		}
	}

	for _, s := range currentResources.Services {
		if err := reports.AddObject(s, nil); err != nil {
			return err
		}
	}

	for _, i := range currentResources.Ingresses {
		if err := reports.AddObject(i, nil); err != nil {
			return err
		}
	}

	for _, cj := range currentResources.CronJobs {
		if err := reports.AddObject(cj, nil); err != nil {
			return err
		}
	}

	return nil
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
	text.SetTitle("Warnings (ESC to exit)")
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
	// output map[string]map[string]*dyff.Report,
	warnings map[string][]*Warning,
) error {
	grid := tview.NewGrid().
		SetRows(0).
		SetColumns(0, 0).
		SetBorders(true).
		AddItem(currentOverview, 0, 0, 10, 1, 0, 0, false).
		AddItem(tview.NewTreeView(), 0, 1, 10, 1, 0, 0, false)

	app := tview.NewApplication().SetRoot(grid, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.Stop()
		}
		return event
	})

	return app.Run()
}

func showDiffReport(r *dyff.Report, kind, name string, warnings []*Warning) error {
	var out bytes.Buffer
	hr := &dyff.HumanReport{
		Report:     *r,
		OmitHeader: true,
	}
	if err := hr.WriteReport(tview.ANSIWriter(&out)); err != nil {
		return err
	}

	text := tview.NewTextView()
	text.SetTitle(fmt.Sprintf("%s/%s (ESC to exit)", kind, name))
	text.SetBorder(true)
	text.SetDynamicColors(true)
	text.SetWrap(true)
	text.SetText(out.String())
	text.SetBackgroundColor(tcell.ColorNone)

	warningsText := getWarningsView(warnings)

	grid := tview.NewGrid().
		SetColumns(0).
		SetBorders(false)

	if warningsText != nil {
		grid.SetRows(10, 0).
			AddItem(warningsText, 0, 0, 1, 1, 0, 0, false).
			AddItem(text, 1, 0, 1, 1, 0, 0, true)
	} else {
		grid.SetRows(0).
			AddItem(text, 0, 0, 1, 1, 0, 0, true)
	}

	app := tview.NewApplication().SetRoot(grid, true).EnableMouse(true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.Stop()
		}
		return event
	})

	return app.Run()
}
