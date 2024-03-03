package migrate

import (
	"encoding/json"

	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// marshall the platform resources into kubernetes resources, and then compare them to the existing k8s resources
func processPlatformOutput(reports map[string]map[string]*dyff.HumanReport,
	currentResources *CurrentResources,
	platformResources map[string]string) error {
	for _, resource := range platformResources {
		// unmarshal the resource into a k8s object
		object := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(resource), object)
		if err != nil {
			return err
		}

		orig, err := currentResources.getCurrentObject(object.GetKind(), object.GetName())
		if err != nil {
			return err
		}

		proposal, err := yaml.YAMLToJSON([]byte(resource))
		if err != nil {
			return err
		}

		b, err := getDiffingReport(orig, proposal)
		if err != nil {
			return err
		}

		if _, ok := reports[object.GetKind()]; !ok {
			reports[object.GetKind()] = map[string]*dyff.HumanReport{}
		}

		reports[object.GetKind()][object.GetName()] = b
	}
	return nil
}

func processOperatorOutput(reports map[string]map[string]*dyff.HumanReport,
	currentResources *CurrentResources,
	operatorOutput []*pipeline.ObjectChange) error {
	for _, out := range operatorOutput {
		orig, err := currentResources.getCurrentObject(out.GetObject().GetGvk().GetKind(), out.GetObject().GetName())
		if err != nil {
			return err
		}

		proposal, err := yaml.YAMLToJSON([]byte(out.GetObject().GetContent()))
		if err != nil {
			return err
		}

		b, err := getDiffingReport(orig, proposal)
		if err != nil {
			return err
		}

		if _, ok := reports[out.GetObject().GetGvk().GetKind()]; !ok {
			reports[out.GetObject().GetGvk().GetKind()] = map[string]*dyff.HumanReport{}
		}

		reports[out.GetObject().GetGvk().GetKind()][out.GetObject().GetName()] = b
	}

	return nil
}

func processRemainingResources(reports map[string]map[string]*dyff.HumanReport,
	currentResources *CurrentResources) error {
	if currentResources.Deployment != nil {
		currentResources.Deployment.SetManagedFields(nil)
		currentResources.Deployment.Status = appsv1.DeploymentStatus{}
		orig, err := json.Marshal(currentResources.Deployment)
		if err != nil {
			return err
		}

		report, err := getDiffingReport(orig, []byte{})
		if err != nil {
			return err
		}

		if _, ok := reports["Deployment"]; !ok {
			reports["Deployment"] = map[string]*dyff.HumanReport{}
		}

		reports["Deployment"][currentResources.Deployment.GetName()] = report
	}

	if currentResources.HPA != nil {
		currentResources.HPA.SetManagedFields(nil)
		orig, err := json.Marshal(currentResources.HPA)
		if err != nil {
			return err
		}

		report, err := getDiffingReport(orig, []byte{})
		if err != nil {
			return err
		}

		if _, ok := reports["HorizontalPodAutoscaler"]; !ok {
			reports["HorizontalPodAutoscaler"] = map[string]*dyff.HumanReport{}
		}

		reports["HorizontalPodAutoscaler"][currentResources.HPA.GetName()] = report
	}

	if currentResources.ServiceAccount != nil {
		currentResources.ServiceAccount.SetManagedFields(nil)
		orig, err := json.Marshal(currentResources.ServiceAccount)
		if err != nil {
			return err
		}

		report, err := getDiffingReport(orig, []byte{})
		if err != nil {
			return err
		}

		if _, ok := reports["ServiceAccount"]; !ok {
			reports["ServiceAccount"] = map[string]*dyff.HumanReport{}
		}

		reports["ServiceAccount"][currentResources.ServiceAccount.GetName()] = report
	}

	if currentResources.Capsule != nil {
		currentResources.Capsule.SetManagedFields(nil)
		currentResources.Capsule.Status = nil
		orig, err := json.Marshal(currentResources.Capsule)
		if err != nil {
			return err
		}

		report, err := getDiffingReport(orig, []byte{})
		if err != nil {
			return err
		}

		if _, ok := reports["Capsule"]; !ok {
			reports["Capsule"] = map[string]*dyff.HumanReport{}
		}

		reports["Capsule"][currentResources.Capsule.GetName()] = report
	}

	for _, cm := range currentResources.ConfigMaps {
		cm.SetManagedFields(nil)
		orig, err := json.Marshal(cm)
		if err != nil {
			return err
		}

		report, err := getDiffingReport(orig, []byte{})
		if err != nil {
			return err
		}

		if _, ok := reports["ConfigMap"]; !ok {
			reports["ConfigMap"] = map[string]*dyff.HumanReport{}
		}

		reports["ConfigMap"][cm.GetName()] = report
	}

	for _, s := range currentResources.Secrets {
		s.SetManagedFields(nil)
		orig, err := json.Marshal(s)
		if err != nil {
			return err
		}

		report, err := getDiffingReport(orig, []byte{})
		if err != nil {
			return err
		}

		if _, ok := reports["Secret"]; !ok {
			reports["Secret"] = map[string]*dyff.HumanReport{}
		}

		reports["Secret"][s.GetName()] = report
	}

	for _, s := range currentResources.Services {
		s.SetManagedFields(nil)
		orig, err := json.Marshal(s)
		if err != nil {
			return err
		}

		report, err := getDiffingReport(orig, []byte{})
		if err != nil {
			return err
		}

		if _, ok := reports["Service"]; !ok {
			reports["Service"] = map[string]*dyff.HumanReport{}
		}

		reports["Service"][s.GetName()] = report
	}

	for _, i := range currentResources.Ingresses {
		i.SetManagedFields(nil)
		orig, err := json.Marshal(i)
		if err != nil {
			return err
		}

		report, err := getDiffingReport(orig, []byte{})
		if err != nil {
			return err
		}

		if _, ok := reports["Ingress"]; !ok {
			reports["Ingress"] = map[string]*dyff.HumanReport{}
		}

		reports["Ingress"][i.GetName()] = report
	}

	for _, cj := range currentResources.CronJobs {
		cj.SetManagedFields(nil)
		orig, err := json.Marshal(cj)
		if err != nil {
			return err
		}

		report, err := getDiffingReport(orig, []byte{})
		if err != nil {
			return err
		}

		if _, ok := reports["CronJob"]; !ok {
			reports["CronJob"] = map[string]*dyff.HumanReport{}
		}

		reports["CronJob"][cj.GetName()] = report
	}

	return nil
}

func getDiffingReport(orig, proposal []byte) (*dyff.HumanReport, error) {
	fromNodes, err := ytbx.LoadDocuments(orig)
	if err != nil {
		return nil, err
	}
	from := ytbx.InputFile{
		Location:  "current",
		Documents: fromNodes,
	}
	toNodes, err := ytbx.LoadDocuments(proposal)
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

	return &dyff.HumanReport{
		Report:     r,
		OmitHeader: true,
	}, nil
}
