package kind

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/helm"
	"github.com/rigdev/rig/pkg/scheme"
	"helm.sh/helm/v3/pkg/strvals"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/kind/pkg/cluster"
	kindexec "sigs.k8s.io/kind/pkg/exec"
)

type ScenarioClient interface {
	Load(path string) (*Scenario, error)
	Run(s *Scenario) error
	Delete(scenarioName string) error
}

func NewScenarioClient(fsys fs.FS) ScenarioClient {
	return &scenarioClient{
		provider: cluster.NewProvider(),
		fs:       fsys,
	}
}

type Scenario struct {
	config *v1alpha1.KindScenarioConfig
	path   string
}

type scenarioClient struct {
	provider *cluster.Provider
	fs       fs.FS
}

func (c *scenarioClient) Delete(scenarioName string) error {
	if err := c.provider.Delete(clusterName(scenarioName), ""); err != nil {
		return fmt.Errorf("could not delete kind cluster: %w", err)
	}
	return nil
}

func (c *scenarioClient) Load(path string) (*Scenario, error) {
	cfgPath := filepath.Join(path, "scenario.yaml")
	bs, err := fs.ReadFile(c.fs, cfgPath)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", cfgPath, err)
	}

	decoder := serializer.NewCodecFactory(scheme.New()).UniversalDeserializer()
	obj, gvk, err := decoder.Decode(bs, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decode %s: %w", cfgPath, err)
	}

	if gvk.Version != v1alpha1.GroupVersion.Version ||
		gvk.Group != v1alpha1.GroupVersion.Group ||
		gvk.Kind != "KindScenarioConfig" {
		return nil, fmt.Errorf("unsupported gvk: %s", gvk.String())
	}

	cfg, ok := obj.(*v1alpha1.KindScenarioConfig)
	if !ok {
		return nil, errors.New("could not assert runtime object to KindScenarioConfig")
	}

	return &Scenario{
		config: cfg,
		path:   path,
	}, nil
}

func (c *scenarioClient) Run(s *Scenario) error {
	if err := c.setupKind(s); err != nil {
		return err
	}

	for _, step := range s.config.Steps {
		if step.Helm != nil {
			if err := c.installHelmRelease(s, step.Helm); err != nil {
				return err
			}
		}

		//if step.Manifest != nil {
		//	if err := s.installManifest(step.Manifest); err != nil {
		//		return err
		//	}
		//}
		//if step.Exec != nil {
		//	if err := s.execInContainer(step.Exec); err != nil {
		//		return err
		//	}
		//}
	}

	return nil
}

func valuesFromEnvVars(mapping map[string]string) (map[string]interface{}, error) {
	base := map[string]interface{}{}
	for path, envVar := range mapping {
		value := fmt.Sprintf("%s=%s", path, os.Getenv(envVar))
		if err := strvals.ParseInto(value, base); err != nil {
			return nil, fmt.Errorf("set value '%s' from envvar '%s': %w", path, envVar, err)
		}
	}
	return base, nil
}

func (c *scenarioClient) installHelmRelease(s *Scenario, helmStep *v1alpha1.KindScenarioHelmStep) error {
	chartName := os.Getenv(helmStep.ChartEnvVar)
	if chartName == "" {
		chartName = helmStep.Chart
	}

	valuesFiles, err := helm.ValuesFromFiles(c.fs, helmStep.ValueFiles...)
	if err != nil {
		return err
	}

	valuesEnv, err := valuesFromEnvVars(helmStep.ValuesFromEnvVars)
	if err != nil {
		return err
	}

	values := helm.MergeValues(valuesFiles, helmStep.Values)
	values = helm.MergeValues(values, valuesEnv)

	hc, err := helm.New(helm.NewClientOptions().
		WithKubeContext(s.kubeCTX()).
		WithNamespace(helmStep.Namespace),
	)
	if err != nil {
		return err
	}

	return hc.Install(
		helmStep.Name,
		chartName,
		helm.NewInstallOptions().
			WithWait(helmStep.Wait).
			WithRepoURL(helmStep.Repo).
			WithVersion(helmStep.Version).
			WithValues(values),
	)
}

func (s *Scenario) kubeCTX() string {
	return fmt.Sprintf("kind-rig-%s", s.config.Name)
}

func (c *scenarioClient) setupKind(s *Scenario) error {
	clusters, err := c.provider.List()
	if err != nil {
		return fmt.Errorf("could not list kind clusters: %w", err)
	}

	clusterName := s.clusterName()
	for _, c := range clusters {
		if c == clusterName {
			return nil
		}
	}

	if err = c.provider.Create(
		clusterName,
		cluster.CreateWithRawConfig([]byte(s.config.KindConfig)),
	); err != nil {
		var rerr *kindexec.RunError
		if errors.As(err, &rerr) {
			return fmt.Errorf("%v: %v", rerr.Inner, string(rerr.Output))
		}
		return err
	}
	return nil
}

const clusterPrefix = "rig"

func (s *Scenario) clusterName() string {
	if s.config == nil {
		return ""
	}
	return clusterName(s.config.Name)
}

func clusterName(scenarioName string) string {
	return fmt.Sprintf("%s-%s", clusterPrefix, scenarioName)
}
