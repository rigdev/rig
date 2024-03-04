package migrate

import (
	"bytes"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rivo/tview"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		currentResources.Deployment.SetManagedFields(nil)
		currentResources.Deployment.Status = appsv1.DeploymentStatus{}
		orig, err := obj.Encode(currentResources.Deployment, scheme)
		if err != nil {
			return err
		}

		report, err := getDiffingReportRaw(orig, nil)
		if err != nil {
			return err
		}

		if _, ok := reports["Deployment"]; !ok {
			reports["Deployment"] = map[string]*dyff.Report{}
		}

		reports["Deployment"][currentResources.Deployment.GetName()] = report
	}

	if currentResources.HPA != nil {
		currentResources.HPA.SetManagedFields(nil)
		orig, err := obj.Encode(currentResources.HPA, scheme)
		if err != nil {
			return err
		}

		report, err := getDiffingReportRaw(orig, nil)
		if err != nil {
			return err
		}

		if _, ok := reports["HorizontalPodAutoscaler"]; !ok {
			reports["HorizontalPodAutoscaler"] = map[string]*dyff.Report{}
		}

		reports["HorizontalPodAutoscaler"][currentResources.HPA.GetName()] = report
	}

	if currentResources.ServiceAccount != nil {
		currentResources.ServiceAccount.SetManagedFields(nil)
		orig, err := obj.Encode(currentResources.ServiceAccount, scheme)
		if err != nil {
			return err
		}

		report, err := getDiffingReportRaw(orig, nil)
		if err != nil {
			return err
		}

		if _, ok := reports["ServiceAccount"]; !ok {
			reports["ServiceAccount"] = map[string]*dyff.Report{}
		}

		reports["ServiceAccount"][currentResources.ServiceAccount.GetName()] = report
	}

	if currentResources.Capsule != nil {
		currentResources.Capsule.SetManagedFields(nil)
		currentResources.Capsule.Status = nil
		orig, err := obj.Encode(currentResources.Capsule, scheme)
		if err != nil {
			return err
		}

		report, err := getDiffingReportRaw(orig, nil)
		if err != nil {
			return err
		}

		if _, ok := reports["Capsule"]; !ok {
			reports["Capsule"] = map[string]*dyff.Report{}
		}

		reports["Capsule"][currentResources.Capsule.GetName()] = report
	}

	for _, cm := range currentResources.ConfigMaps {
		cm.SetManagedFields(nil)
		orig, err := obj.Encode(cm, scheme)
		if err != nil {
			return err
		}

		report, err := getDiffingReportRaw(orig, nil)
		if err != nil {
			return err
		}

		if _, ok := reports["ConfigMap"]; !ok {
			reports["ConfigMap"] = map[string]*dyff.Report{}
		}

		reports["ConfigMap"][cm.GetName()] = report
	}

	for _, s := range currentResources.Secrets {
		s.SetManagedFields(nil)
		orig, err := obj.Encode(s, scheme)
		if err != nil {
			return err
		}

		report, err := getDiffingReportRaw(orig, nil)
		if err != nil {
			return err
		}

		if _, ok := reports["Secret"]; !ok {
			reports["Secret"] = map[string]*dyff.Report{}
		}

		reports["Secret"][s.GetName()] = report
	}

	for _, s := range currentResources.Services {
		s.SetManagedFields(nil)
		orig, err := obj.Encode(s, scheme)
		if err != nil {
			return err
		}

		report, err := getDiffingReportRaw(orig, nil)
		if err != nil {
			return err
		}

		if _, ok := reports["Service"]; !ok {
			reports["Service"] = map[string]*dyff.Report{}
		}

		reports["Service"][s.GetName()] = report
	}

	for _, i := range currentResources.Ingresses {
		i.SetManagedFields(nil)
		orig, err := obj.Encode(i, scheme)
		if err != nil {
			return err
		}

		report, err := getDiffingReportRaw(orig, nil)
		if err != nil {
			return err
		}

		if _, ok := reports["Ingress"]; !ok {
			reports["Ingress"] = map[string]*dyff.Report{}
		}

		reports["Ingress"][i.GetName()] = report
	}

	for _, cj := range currentResources.CronJobs {
		cj.SetManagedFields(nil)
		orig, err := obj.Encode(cj, scheme)
		if err != nil {
			return err
		}

		report, err := getDiffingReportRaw(orig, nil)
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
	if err := normalize(orig, scheme); err != nil {
		return nil, err
	}
	if err := normalize(proposal, scheme); err != nil {
		return nil, err
	}
	trimAnnotations(orig)
	trimAnnotations(proposal)

	origBytes, err := obj.Encode(orig, scheme)
	if err != nil {
		return nil, err
	}

	proposalBytes, err := obj.Encode(proposal, scheme)
	if err != nil {
		return nil, err
	}

	return getDiffingReportRaw(origBytes, proposalBytes)
}

func getDiffingReportRaw(orig, proposal []byte) (*dyff.Report, error) {
	if len(orig) == 0 {
		orig = []byte("{}")
	}
	if len(proposal) == 0 {
		proposal = []byte("{}")
	}

	fromNodes, err := ytbx.LoadYAMLDocuments(orig)
	if err != nil {
		return nil, err
	}
	from := ytbx.InputFile{
		Location:  "current",
		Documents: fromNodes,
	}
	toNodes, err := ytbx.LoadYAMLDocuments(proposal)
	if err != nil {
		return nil, err
	}
	to := ytbx.InputFile{
		Location:  "migration",
		Documents: toNodes,
	}

	r, err := dyff.CompareInputFiles(from, to)
	if err != nil {
		return nil, err
	}

	// for i, d := range r.Diffs {
	// 	fmt.Println(d.Path.ToDotStyle())
	// 	if d.Path == nil {
	// 		continue
	// 	}
	// 	if strings.HasPrefix(d.Path.ToDotStyle(), "status") {
	// 		// fmt.Println(d.Path.ToDotStyle())
	// 		r.Diffs = append(r.Diffs[:i], r.Diffs[i+1:]...)
	// 	}
	// }

	return &r, nil
}

func trimAnnotations(co client.Object) {
	if co == nil {
		return
	}

	annotations := co.GetAnnotations()
	if annotations == nil {
		return
	}

	delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
	delete(annotations, "deployment.kubernetes.io/revision")

	co.SetAnnotations(annotations)
}

func normalize(co client.Object, scheme *runtime.Scheme) error {
	if co == nil {
		return nil
	}

	gvks, _, err := scheme.ObjectKinds(co)
	if err != nil {
		return err
	}

	co.GetObjectKind().SetGroupVersionKind(gvks[0])
	co.SetManagedFields(nil)
	co.SetCreationTimestamp(v1.Time{})
	co.SetGeneration(0)
	co.SetResourceVersion("")
	co.SetOwnerReferences(nil)
	co.SetUID("")
	return nil
}

func showDiffReport(r *dyff.Report, kind, name string) error {
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

	app := tview.NewApplication().SetRoot(text, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			app.Stop()
		}
		return event
	})

	return app.Run()
}
