package pipeline

import (
	"github.com/rigdev/rig/pkg/obj"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AppsDeploymentGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")

func ObjectsEquals(o1, o2 client.Object, scheme *runtime.Scheme) (bool, error) {
	comp := obj.NewComparison(o1, o2, scheme)

	// Ignore deployment-revision. This is added in a post-step by the deployment step.
	comp.AddFilter(obj.RemoveAnnotationsFilter(
		"deployment.kubernetes.io/revision",
	))

	// Always ignore `status` properties.
	comp.AddRemoveDiffs("status")

	diff, err := comp.ComputeDiff()
	if err != nil {
		return false, err
	}

	// Out-comment this for pretty-printing the diff between the files.
	/*
		hr := &dyff.HumanReport{
			Report:     *diff.Report,
			OmitHeader: true,
		}

		if err := hr.WriteReport(os.Stderr); err != nil {
			panic(err)
		}
		// */

	return len(diff.Report.Diffs) == 0, nil
}
