package kind

import (
	"context"
	"embed"
	"fmt"

	"github.com/fatih/color"
	"github.com/rigdev/rig/pkg/kind"
	"github.com/spf13/cobra"
)

//go:embed scenarios
var scenariosFS embed.FS

func (c *Cmd) scenarioCreate(_ context.Context, _ *cobra.Command, _ []string) error {
	sc := kind.NewScenarioClient(scenariosFS)

	s, err := sc.Load("scenarios/default")
	if err != nil {
		return err
	}

	return sc.Run(s)
}

func (c *Cmd) scenarioClean(_ context.Context, _ *cobra.Command, _ []string) error {
	name := "default"
	fmt.Printf("Deleting scenario %s...", name)
	if err := kind.NewScenarioClient(scenariosFS).Delete(name); err != nil {
		return err
	}
	color.Green(" ✓")
	return nil
}

/* type scenario struct {
	name         string
	fs           fs.FS
	scenarioDir  string
	config       *v1alpha1.KindScenarioConfig
	dockerClient *dockerclient.Client
}
*/
/* func (c Cmd) loadScenario(name string) (*scenario, error) {
	e := &scenario{
		name:         name,
		fs:           scenariosFS, // TODO: support using real fs
		scenarioDir:  filepath.Join("scenarios", name),
		dockerClient: c.DockerClient,
	}

	bs, err := fs.ReadFile(e.fs, filepath.Join(e.scenarioDir, "scenario.yaml"))
	if err != nil {
		return nil, fmt.Errorf("could not load scenario.yaml for scenario %s: %w", name, err)
	}

	decoder := serializer.NewCodecFactory(scheme.New()).UniversalDeserializer()
	obj, gvk, err := decoder.Decode(bs, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decode scenario.yaml for scenario %s: %w", name, err)
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

	e.config = cfg

	return e, nil
} */

/* func (s *scenario) create(ctx context.Context) error {
	if err := s.setupKind(); err != nil {
		return err
	}

	for _, step := range s.config.Steps {
		if step.Helm != nil {
			if err := s.installHelmRelease(ctx, step.Helm); err != nil {
				return err
			}
		}
		if step.Manifest != nil {
			if err := s.installManifest(step.Manifest); err != nil {
				return err
			}
		}
		if step.Exec != nil {
			if err := s.execInContainer(step.Exec); err != nil {
				return err
			}
		}
	}

	return nil
} */

/* func (s *scenario) execInContainer(execStep *v1alpha1.KindScenarioExecStep) error {
	args := []string{
		"exec", "--namespace", execStep.Namespace,
	}

	if execStep.TTY {
		args = append(args, "--tty")
	}
	if execStep.Stdin {
		args = append(args, "--stdin")
	}
	if execStep.Container != "" {
		args = append(args, "--container", execStep.Container)
	}

	args = append(args, execStep.Reference, "--")
	args = append(args, execStep.Command...)

	desc := fmt.Sprintf(
		"Executing command `%s` in %s...",
		strings.Join(execStep.Command, " "),
		execStep.Reference,
	)

	fmt.Println(desc)

	cmd := exec.Command(
		"kubectl", args...,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not execute command: %w", err)
	}
	return nil
}

func (s *scenario) clusterName() string {
	return fmt.Sprintf("rig-%s", s.name)
}

func (s *scenario) kubeContext() string {
	return fmt.Sprintf("kind-%s", s.clusterName())
} */

/* func (s *scenario) getValues(releaseName, releaseNamespace string) ([]byte, error) {
	fileName := filepath.Join(s.scenarioDir, "values", releaseNamespace, fmt.Sprintf("%s.yaml", releaseName))
	if _, err := fs.Stat(s.fs, fileName); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not stat file %s: %w", fileName, err)
	}
	bs, err := fs.ReadFile(s.fs, fileName)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", fileName, err)
	}
	return bs, nil
} */

/* func (s *scenario) setupKind() error {
	var err error
	fmt.Printf("Creating kind cluster '%s' if not present...", s.clusterName())
	defer func() {
		if err == nil {
			color.Green(" ✓")
		}
	}()

	provider := cluster.NewProvider()
	clusters, err := provider.List()
	if err != nil {
		return err
	}

	clusterName := s.clusterName()
	if slices.Contains(clusters, clusterName) {
		return nil
	}

	if err = provider.Create(
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
} */

/* const (
	rigHelmRepo                = "https://charts.rig.dev"
	rigOperatorChart           = "rig-operator"
	rigPlatformChart           = "rig-platform"
	rigOperatorChartPathEnvVar = "RIG_OPERATOR_CHART_PATH"
	rigPlatformChartPathEnvVar = "RIG_PLATFORM_CHART_PATH"
	rigOperatorDockerTagEnvVar = "RIG_OPERATOR_DOCKER_TAG"
	rigPlatformDockerTagEnvVar = "RIG_PLATFORM_DOCKER_TAG"
	rigOperatorDockerRepo      = "ghcr.io/rigdev/rig-operator"
	rigPlatformDockerRepo      = "ghcr.io/rigdev/rig-platform"
) */

/* func (s *scenario) loadImage(ctx context.Context, image string) error {
	res, err := s.dockerClient.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: image,
		}),
	})
	if err != nil {
		return fmt.Errorf("could not list docker images")
	}

	if len(res) == 0 || imageTag(image) == "latest" {
		if err := runCmd(
			fmt.Sprintf("Pulling docker image %s", image),
			"docker", "pull", image,
		); err != nil {
			return err
		}
	}

	if err := runCmd(
		fmt.Sprintf("Loading docker image %s into kind cluster %s...", image, s.clusterName()),
		kindBin.bin(), "load", "docker-image", image, "-n", s.clusterName(),
	); err != nil {
		return fmt.Errorf("could not load docker image %s into kind cluster %s: %w", image, s.clusterName(), err)
	}

	return nil
} */

/* func imageTag(img string) string {
	parts := strings.Split(img, ":")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
} */

/* func (s *scenario) installHelmRelease(ctx context.Context, helmStep *v1alpha1.KindScenarioHelmStep) error {
	if s.config == nil {
		return nil
	}

	args := []string{
		"--kube-context", s.kubeContext(),
		"upgrade", "--install",
		helmStep.Name,
	}

	if helmStep.Repo == rigHelmRepo {
		if helmStep.Chart == rigOperatorChart {
			if chartPath := os.Getenv(rigOperatorChartPathEnvVar); chartPath != "" {
				args = append(args, chartPath)
				helmStep.Repo = ""
			} else {
				args = append(args, helmStep.Chart)
			}

			if dockerTag := os.Getenv(rigOperatorDockerTagEnvVar); dockerTag != "" {
				img := fmt.Sprintf("%s:%s", rigOperatorDockerRepo, dockerTag)
				if err := s.loadImage(ctx, img); err != nil {
					return err
				}
				args = append(args, "--set", fmt.Sprintf("image.tag=%s", dockerTag))
			}
		}

		if helmStep.Chart == rigPlatformChart {
			if chartPath := os.Getenv(rigPlatformChartPathEnvVar); chartPath != "" {
				args = append(args, chartPath)
				helmStep.Repo = ""
			} else {
				args = append(args, helmStep.Chart)
			}

			if dockerTag := os.Getenv(rigPlatformDockerTagEnvVar); dockerTag != "" {
				img := fmt.Sprintf("%s:%s", rigPlatformDockerRepo, dockerTag)
				if err := s.loadImage(ctx, img); err != nil {
					return err
				}
				args = append(args, "--set", fmt.Sprintf("image.tag=%s", dockerTag))
			}
		}
	} else {
		args = append(args, helmStep.Chart)
	}

	args = append(args, "--namespace", helmStep.Namespace, "--create-namespace")

	if helmStep.Repo != "" {
		args = append(args, "--repo", helmStep.Repo)
	}

	if helmStep.Version != "" && helmStep.Repo != "" {
		args = append(args, "--version", helmStep.Version)
	}

	if helmStep.Wait {
		args = append(args, "--wait")
	}

	vals, err := s.getValues(helmStep.Name, helmStep.Namespace)
	if err != nil {
		return err
	}
	hasValues := len(vals) > 0

	if hasValues {
		if vals != nil {
			args = append(args, "-f", "-")
		}
	}

	cmd := common.NewDeferredOutputCommand(fmt.Sprintf(
		"Installing helm release %s/%s...", helmStep.Namespace, helmStep.Name,
	))
	cmd.Command("helm", args...)

	var stdin io.WriteCloser
	if hasValues {
		stdin, err = cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("could not get stdin pipe: %w", err)
		}
		defer stdin.Close()
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start command: %w", err)
	}

	if hasValues {
		if _, err := stdin.Write(vals); err != nil {
			return fmt.Errorf("could not pipe helm values: %w", err)
		}
		stdin.Close()
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error when waiting for command: %w", err)
	}

	cmd.End(true)

	return nil
} */

/* func (s *scenario) installManifest(manifestStep *v1alpha1.KindScenarioManifestStep) error {
	manifestPath := filepath.Join(s.scenarioDir, manifestStep.Path)
	_, err := fs.Stat(s.fs, manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("could not stat manifest dir %s: %w", manifestPath, err)
	}

	err = fs.WalkDir(s.fs, manifestPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("could not walk path %s: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		fileExt := filepath.Ext(d.Name())
		if fileExt != ".yaml" && fileExt != ".yml" {
			return nil
		}

		bs, err := fs.ReadFile(s.fs, path)
		if err != nil {
			return fmt.Errorf("could not read manifest file %s: %w", path, err)
		}

		cmd := common.NewDeferredOutputCommand(fmt.Sprintf("Applying manifests from %s...", path))
		cmd.Command("kubectl", "apply", "-f", "-")

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("could not get stdin pipe for kubectl: %w", err)
		}
		defer stdin.Close()

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("could not start kubectl command: %w", err)
		}

		if _, err := stdin.Write(bs); err != nil {
			return fmt.Errorf("could not write to kubectl pipe: %w", err)
		}
		stdin.Close()

		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("error when waiting for kubectl command: %w", err)
		}

		cmd.End(true)
		return nil
	})
	return err
} */
