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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// marshall the platform resources into kubernetes resources, and then compare them to the existing k8s resources
func processPlatformOutput(reports map[string]map[string]*dyff.Report,
	currentResources *CurrentResources,
	platformResources map[string]string,
	scheme *runtime.Scheme,
) (*v1alpha2.Capsule, error) {
	var capsule *v1alpha2.Capsule
	for _, resource := range platformResources {
		// unmarshal the resource into a k8s object
		object := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(resource), object)
		if err != nil {
			return nil, err
		}

		orig := currentResources.getCurrentObject(object.GetKind(), object.GetName())

		proposal, err := obj.DecodeAny([]byte(resource), scheme)
		if err != nil {
			return nil, err
		}

		b, err := getDiffingReport(orig, proposal, scheme)
		if err != nil {
			return nil, err
		}
		if _, ok := reports[object.GetKind()]; !ok {
			if object.GetKind() == "Capsule" {
				capsule = &v1alpha2.Capsule{}
				err = obj.Decode([]byte(resource), capsule)
				if err != nil {
					return nil, err
				}
			}
			reports[object.GetKind()] = map[string]*dyff.Report{}
		}

		reports[object.GetKind()][object.GetName()] = b
	}

	return capsule, nil
}

func processOperatorOutput(reports map[string]map[string]*dyff.Report,
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

		b, err := getDiffingReport(orig, proposal, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports[out.GetObject().GetGvk().GetKind()]; !ok {
			reports[out.GetObject().GetGvk().GetKind()] = map[string]*dyff.Report{}
		}

		reports[out.GetObject().GetGvk().GetKind()][out.GetObject().GetName()] = b
	}

	return nil
}

func processRemainingResources(reports map[string]map[string]*dyff.Report,
	currentResources *CurrentResources,
	scheme *runtime.Scheme,
) error {
	if currentResources.Deployment != nil {
		report, err := getDiffingReport(currentResources.Deployment, nil, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports["Deployment"]; !ok {
			reports["Deployment"] = map[string]*dyff.Report{}
		}

		reports["Deployment"][currentResources.Deployment.GetName()] = report
	}

	if currentResources.HPA != nil {
		report, err := getDiffingReport(currentResources.HPA, nil, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports["HorizontalPodAutoscaler"]; !ok {
			reports["HorizontalPodAutoscaler"] = map[string]*dyff.Report{}
		}

		reports["HorizontalPodAutoscaler"][currentResources.HPA.GetName()] = report
	}

	if currentResources.ServiceAccount != nil {
		report, err := getDiffingReport(currentResources.ServiceAccount, nil, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports["ServiceAccount"]; !ok {
			reports["ServiceAccount"] = map[string]*dyff.Report{}
		}

		reports["ServiceAccount"][currentResources.ServiceAccount.GetName()] = report
	}

	if currentResources.Capsule != nil {
		report, err := getDiffingReport(currentResources.Capsule, nil, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports["Capsule"]; !ok {
			reports["Capsule"] = map[string]*dyff.Report{}
		}

		reports["Capsule"][currentResources.Capsule.GetName()] = report
	}

	for _, cm := range currentResources.ConfigMaps {
		report, err := getDiffingReport(cm, nil, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports["ConfigMap"]; !ok {
			reports["ConfigMap"] = map[string]*dyff.Report{}
		}

		reports["ConfigMap"][cm.GetName()] = report
	}

	for _, s := range currentResources.Secrets {
		report, err := getDiffingReport(s, nil, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports["Secret"]; !ok {
			reports["Secret"] = map[string]*dyff.Report{}
		}

		reports["Secret"][s.GetName()] = report
	}

	for _, s := range currentResources.Services {
		report, err := getDiffingReport(s, nil, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports["Service"]; !ok {
			reports["Service"] = map[string]*dyff.Report{}
		}

		reports["Service"][s.GetName()] = report
	}

	for _, i := range currentResources.Ingresses {
		report, err := getDiffingReport(i, nil, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports["Ingress"]; !ok {
			reports["Ingress"] = map[string]*dyff.Report{}
		}

		reports["Ingress"][i.GetName()] = report
	}

	for _, cj := range currentResources.CronJobs {
		report, err := getDiffingReport(cj, nil, scheme)
		if err != nil {
			return err
		}

		if _, ok := reports["CronJob"]; !ok {
			reports["CronJob"] = map[string]*dyff.Report{}
		}

		reports["CronJob"][cj.GetName()] = report
	}

	return nil
}

func getDiffingReport(orig, proposal client.Object, scheme *runtime.Scheme) (*dyff.Report, error) {
	c := obj.NewComparison(orig, proposal, scheme)
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

func showOverview(currentOverview *tview.TreeView,
	output map[string]map[string]*dyff.Report,
	warnings map[string][]*Warning) error {
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
