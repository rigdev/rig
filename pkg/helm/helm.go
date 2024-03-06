package helm

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type Client interface {
	Install(releaseName, chart string, opts *InstallOptions) error
}

func New(opts *ClientOptions) (Client, error) {
	settings := cli.New()

	if opts.KubeContext != "" {
		settings.KubeContext = opts.KubeContext
	}
	if opts.Namespace != "" {
		settings.SetNamespace(opts.Namespace)
	}

	actionCFG := &action.Configuration{}
	if err := actionCFG.Init(
		settings.RESTClientGetter(),
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		log.Printf,
	); err != nil {
		return nil, fmt.Errorf("could not init helm action config: %w", err)
	}

	return &client{
		settings:  settings,
		actionCFG: actionCFG,
	}, nil
}

type ClientOptions struct {
	KubeContext string
	Namespace   string
}

func NewClientOptions() *ClientOptions {
	return &ClientOptions{}
}

func (opts *ClientOptions) WithKubeContext(ctx string) *ClientOptions {
	opts.KubeContext = ctx
	return opts
}

func (opts *ClientOptions) WithNamespace(namespace string) *ClientOptions {
	opts.Namespace = namespace
	return opts
}

type client struct {
	settings  *cli.EnvSettings
	actionCFG *action.Configuration
}

type InstallOptions struct {
	RepoURL string
	Version string
	Values  map[string]interface{}
	Wait    bool
}

func NewInstallOptions() *InstallOptions {
	return &InstallOptions{}
}

type InstallOption func(*InstallOptions)

func (opts *InstallOptions) WithWait(wait bool) *InstallOptions {
	opts.Wait = wait
	return opts
}

func (opts *InstallOptions) WithVersion(version string) *InstallOptions {
	opts.Version = version
	return opts
}

func (opts *InstallOptions) WithRepoURL(repo string) *InstallOptions {
	opts.RepoURL = repo
	return opts
}

func (opts *InstallOptions) WithValues(values map[string]interface{}) *InstallOptions {
	opts.Values = values
	return opts
}

func (c *client) Install(
	releaseName string,
	chart string,
	opts *InstallOptions,
) error {
	if opts == nil {
		opts = NewInstallOptions()
	}

	client := action.NewUpgrade(c.actionCFG)
	client.Namespace = c.settings.Namespace()
	client.Wait = opts.Wait
	client.Timeout = time.Minute * 5
	client.RepoURL = opts.RepoURL
	client.Version = opts.Version

	chartPath, err := client.ChartPathOptions.LocateChart(chart, c.settings)
	if err != nil {
		return fmt.Errorf("coult not locate helm chart: %w", err)
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("could not load chart: %w", err)
	}

	histClient := action.NewHistory(c.actionCFG)
	histClient.Max = 1
	if _, err := histClient.Run(releaseName); err == driver.ErrReleaseNotFound {
		instClient := action.NewInstall(c.actionCFG)
		instClient.ReleaseName = releaseName
		instClient.CreateNamespace = true

		instClient.Namespace = c.settings.Namespace()
		instClient.Wait = opts.Wait
		instClient.Timeout = time.Minute * 5
		instClient.RepoURL = opts.RepoURL
		instClient.Version = opts.Version

		if _, err := instClient.Run(chartRequested, opts.Values); err != nil {
			return fmt.Errorf("could not helm install: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("could not get helm release history: %w", err)
	}

	_, err = client.Run(releaseName, chartRequested, opts.Values)
	if err != nil {
		return fmt.Errorf("could not run helm install: %w", err)
	}

	return nil
}

func ValuesFromFiles(fsys fs.FS, files ...string) (map[string]interface{}, error) {
	base := map[string]interface{}{}
	for _, f := range files {
		file, err := fsys.Open(f)
		if err != nil {
			return nil, fmt.Errorf("could not open value file %s: %w", f, err)
		}
		values := map[string]interface{}{}
		if err := yaml.NewDecoder(file).Decode(&values); err != nil {
			return nil, fmt.Errorf("could not decode values: %w", err)
		}
		base = MergeValues(base, values)
	}
	return base, nil
}

func MergeValues(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeValues(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
