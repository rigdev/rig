package migrate

import (
	"strings"

	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/roclient"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createHelmReader(scheme *runtime.Scheme, helmDir string, valuesFiles []string) (client.Reader, error) {
	chart, err := loader.Load(helmDir)
	if err != nil {
		return nil, err
	}

	notesIndex := -1
	for i, t := range chart.Templates {
		if strings.Contains(t.Name, "NOTES.txt") {
			notesIndex = i
			break
		}
	}

	if notesIndex != -1 {
		chart.Templates = append(chart.Templates[:notesIndex], chart.Templates[notesIndex+1:]...)
	}

	vals := chart.Values
	for _, valuesFile := range valuesFiles {
		fileValues, err := chartutil.ReadValuesFile(valuesFile)
		if err != nil {
			return nil, err
		}

		vals = chartutil.CoalesceTables(fileValues, vals)
	}

	releaseOpts := chartutil.ReleaseOptions{
		Name:      "migrate",
		Namespace: "migrate",
		Revision:  1,
		IsInstall: true,
	}
	valuesToRender, err := chartutil.ToRenderValues(chart, vals, releaseOpts, nil)
	if err != nil {
		return nil, err
	}

	cfg, err := base.GetRestConfig()
	if err != nil {
		return nil, err
	}
	eng := engine.New(cfg)
	out, err := eng.Render(chart, valuesToRender)
	if err != nil {
		return nil, err
	}

	objs, err := ProcessHelmOutput(out, scheme)
	if err != nil {
		return nil, err
	}

	reader := roclient.NewReader(scheme)
	for _, obj := range objs {
		if err := reader.AddObject(obj); err != nil {
			return nil, err
		}
	}

	return reader, nil
}
