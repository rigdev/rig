package kind

import (
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/fatih/color"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-operator/certgen"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kind/pkg/cluster"
	kindexec "sigs.k8s.io/kind/pkg/exec"
)

//go:embed config.yaml
var config string

//go:embed registry.yaml
var registry string

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, args []string) error {
	if err := checkBinaries(kubectl, kind, helm); err != nil {
		return err
	}

	if err := setupKindRigCluster(); err != nil {
		return err
	}

	images := map[string]string{
		"postgres": "16",
	}

	go func() {
		deferred := common.NewDeferredOutputCommand("")
		for image, tag := range images {
			if err := c.loadImage(ctx, deferred, image, tag); err != nil {
				fmt.Printf("Warning; error fetching image '%s:%s': %v\n", image, tag, err)
			}
		}
	}()

	if err := setupK8s(); err != nil {
		return err
	}

	if err := helmInstall(); err != nil {
		return err
	}

	if err := c.deploy(ctx, cmd, args); err != nil {
		return err
	}

	fmt.Print("Waiting for rig-platform to be ready...")

	hc := http.Client{}
	hc.Timeout = 2 * time.Second
	n := 0
	for {
		res, err := hc.Get("http://localhost:4747/")
		if err != nil {
			n++
			if n > 200 {
				return err
			}

			time.Sleep(500 * time.Millisecond)
			continue
		}

		res.Body.Close()
		break
	}
	color.Green(" ✓")

	if !skipInit {
		fmt.Println()
		fmt.Println("To use Rig you need to create at least one admin user.")
		if err := c.runInit(cmd, args); err != nil {
			return err
		}
	}
	fmt.Println("Rig is now accessible on http://localhost:4747")

	return nil
}

func (c *Cmd) deploy(ctx context.Context, _ *cobra.Command, _ []string) error {
	if err := checkBinaries(kind, kubectl, helm, docker); err != nil {
		return err
	}

	if operatorDockerTag == "" {
		operatorDockerTag = "latest"
	}

	cert, err := certgen.GenerateCerts([]string{
		"rig-operator",
		"rig-operator.rig-system.svc",
	})
	if err != nil {
		return err
	}

	version := time.Now().UnixMilli()

	operatorArgs := []string{
		"--set", fmt.Sprintf("image.tag=%s", operatorDockerTag),
		"--set", fmt.Sprintf("podAnnotations.rig\\.dev\\.reload=ts%d", version),
		"--set", fmt.Sprintf("certgen.certificate.ca=%s", base64.StdEncoding.EncodeToString(cert.CA)),
		"--set", fmt.Sprintf("certgen.certificate.cert=%s", base64.StdEncoding.EncodeToString(cert.Cert)),
		"--set", fmt.Sprintf("certgen.certificate.key=%s", base64.StdEncoding.EncodeToString(cert.Key)),
		"--set", "apicheck.enabled=false",
	}
	if prometheus {
		operatorArgs = append(operatorArgs, "--set", "config.prometheusServiceMonitor.portName=metrics")
	}
	if vpa {
		operatorArgs = append(operatorArgs, "--set", "config.verticalPodAutoscaler.enabled=true")
	}
	if operatorValues != "" {
		operatorArgs = append(operatorArgs, "--values", operatorValues)
	}

	if err := c.deployInner(ctx, deployParams{
		dockerImage: "ghcr.io/rigdev/rig-operator",
		dockerTag:   operatorDockerTag,
		chartName:   "rig-operator",
		chartPath:   operatorChartPath,
		customArgs:  operatorArgs,
	}); err != nil {
		return err
	}

	if platformDockerTag == "" {
		platformDockerTag = "latest"
	}
	platformArgs := []string{
		"--set", fmt.Sprintf("image.tag=%s", platformDockerTag),
		"--set", "rig.clusters.kind.type=k8s",
		"--set", "rig.clusters.kind.devRegistry.host=localhost:30000",
		"--set", "rig.clusters.kind.devRegistry.clusterHost=registry:5000",
		"--set", "rig.environments.kind.cluster=kind",
		"--set", "postgres.enabled=true",
		"--set", "loadBalancer.enabled=true",
		"--set", fmt.Sprintf("rollout=%d", version),
	}
	if platformValues != "" {
		platformArgs = append(platformArgs, "--values", platformValues)
	}

	if err := c.deployInner(ctx, deployParams{
		dockerImage: "ghcr.io/rigdev/rig-platform",
		dockerTag:   platformDockerTag,
		chartName:   "rig-platform",
		chartPath:   platformChartPath,
		customArgs:  platformArgs,
		rollout:     fmt.Sprint(version),
	}); err != nil {
		return err
	}

	return nil
}

func waitUntilDeploymentIsReady(
	cmd *common.DeferredOutputCommand,
	deployment string,
	rollout string,
) error {
	type ready struct {
		Spec struct {
			Template struct {
				Metadata struct {
					Annotations map[string]string
				}
			}
		}
		Status struct {
			Replicas            int `yaml:"replicas,omitempty"`
			UnavailableReplicas int `yaml:"unavailableReplicas,omitempty"`
			AvailableReplicas   int `yaml:"availableReplicas,omitempty"`
			UpdatedReplicas     int `yaml:"updatedReplicas,omitempty"`
			ReadyReplicas       int `yaml:"readyReplicas,omitempty"`
		} `yaml:"status,omitempty"`
	}
	c := 0
	for {
		out, err := cmd.Command(
			"kubectl", "--context", "kind-rig", "get", deployment, "-n", "rig-system", "-oyaml",
		).Output()
		if err != nil {
			c++
			if c > 50 {
				return err
			}

			time.Sleep(500 * time.Millisecond)
			continue
		}

		var r ready
		if err := yaml.Unmarshal(out, &r); err != nil {
			return err
		}

		currentRollout := r.Spec.Template.Metadata.Annotations["rig.dev/rollout"]

		if currentRollout == rollout &&
			r.Status.Replicas >= 1 &&
			r.Status.AvailableReplicas == r.Status.Replicas &&
			r.Status.UpdatedReplicas == r.Status.Replicas {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	return nil
}

type deployParams struct {
	dockerImage string
	dockerTag   string
	chartName   string
	chartPath   string
	customArgs  []string
	rollout     string
}

func (c *Cmd) deployInner(ctx context.Context, p deployParams) error {
	var err error
	cmd := common.NewDeferredOutputCommand(fmt.Sprintf("Deploying %s...", p.chartName))
	defer func() {
		cmd.End(err == nil)
	}()

	if err = c.loadImage(ctx, cmd, p.dockerImage, p.dockerTag); err != nil {
		return err
	}
	chart := p.chartName
	if p.chartPath != "" {
		chart = p.chartPath
	}
	cArgs := []string{
		"--kube-context", "kind-rig",
		"upgrade", "--install", p.chartName, chart,
		"--namespace", "rig-system",
		"--set", fmt.Sprintf("image.tag=%s", p.dockerTag),
		"--create-namespace",
	}
	cArgs = append(cArgs, p.customArgs...)
	if p.chartPath == "" {
		cArgs = append(cArgs, "--repo", "https://charts.rig.dev")
	}
	if err = cmd.RunNew("helm", cArgs...); err != nil {
		return err
	}

	if err = waitUntilDeploymentIsReady(cmd, fmt.Sprintf("deployment.apps/%s", p.chartName), p.rollout); err != nil {
		return err
	}

	return nil
}

func (c *Cmd) loadImage(ctx context.Context, cmd *common.DeferredOutputCommand, img, tag string) error {
	imageTag := fmt.Sprintf("%s:%s", img, tag)
	res, err := c.DockerClient.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: imageTag,
		}),
	})
	if err != nil {
		return err
	}
	if len(res) == 0 || tag == "latest" {
		if err := cmd.RunNew("docker", "pull", imageTag); err != nil {
			return err
		}
	}

	if err := cmd.RunNew("kind", "load", "docker-image", imageTag, "-n", "rig"); err != nil {
		return err
	}

	return nil
}

func (c *Cmd) clean(_ context.Context, _ *cobra.Command, _ []string) error {
	if err := checkBinaries(kind); err != nil {
		return err
	}

	if err := runCmd("Cleaning Rig kind cluster...", "kind", "delete", "clusters", "rig"); err != nil {
		return err
	}

	return nil
}

func setupKindRigCluster() error {
	var err error
	fmt.Print("Creating kind cluster 'rig' if not present...")
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

	if slices.Contains(clusters, "rig") {
		return nil
	}

	if err = provider.Create(
		"rig",
		cluster.CreateWithRawConfig([]byte(config)),
	); err != nil {
		var rerr *kindexec.RunError
		if errors.As(err, &rerr) {
			return fmt.Errorf("%v: %v", rerr.Inner, string(rerr.Output))
		}
		return err
	}

	return nil
}

func setupK8s() error {
	var err error
	cmd := common.NewDeferredOutputCommand("Setup kind cluster...", common.Verbose(verbose))
	defer func() {
		cmd.End(err == nil)
	}()

	if err = cmd.RunNew("kubectl", "--context", "kind-rig", "get", "namespace", "rig-system"); err != nil {
		if err = cmd.RunNew("kubectl", "--context", "kind-rig", "create", "namespace", "rig-system"); err != nil {
			return err
		}
	}

	cmd.Command("kubectl", "--context", "kind-rig", "apply", "-n", "rig-system", "-f", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	if err = cmd.Start(); err != nil {
		return err
	}

	if _, err := io.WriteString(stdin, registry); err != nil {
		return fmt.Errorf("could not write registry to kubectl command: %w", err)
	}
	stdin.Close()
	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func helmInstall() error {
	if prometheus {
		cmd := common.NewDeferredOutputCommand("Installing prometheus...", common.Verbose(verbose))
		cmd.Command(
			"helm", "--kube-context", "kind-rig",
			"upgrade", "--install", "kube-prometheus-stack", "kube-prometheus-stack",
			"--repo", "https://prometheus-community.github.io/helm-charts",
			"--create-namespace",
			"--namespace", "prometheus",
			"-f", "-",
		)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		defer stdin.Close()
		value := `
prometheus:
  prometheusSpec:
    serviceMonitorSelector:
      matchExpressions:
        - key: rig.dev/capsule
          operator: Exists
`
		if err := cmd.Start(); err != nil {
			return err
		}
		_, err = stdin.Write([]byte(value))
		if err != nil {
			return err
		}
		stdin.Close()
		if err := cmd.Wait(); err != nil {
			return err
		}
		cmd.End(true)

		if err := runCmd("Installing prometheus-adapter...",
			"helm", "--kube-context", "kind-rig",
			"upgrade", "--install", "prometheus-adapter", "prometheus-adapter",
			"--repo", "https://prometheus-community.github.io/helm-charts",
			"--namespace", "prometheus",
			"--set", "prometheus.url=http://kube-prometheus-stack-prometheus.prometheus.svc",
		); err != nil {
			return err
		}
	}

	if err := runCmd("Installing metrics-server...",
		"helm", "--kube-context", "kind-rig",
		"upgrade", "--install", "metrics-server", "metrics-server",
		"--repo", "https://kubernetes-sigs.github.io/metrics-server",
		"--namespace", "kube-system",
		"--set", "args={--kubelet-insecure-tls}",
	); err != nil {
		return err
	}

	if vpa {
		if err := runCmd("Installing Vertical Pod Autoscaler CRD", "kubectl", "apply", "-f",
			"https://raw.githubusercontent.com/kubernetes/autoscaler/master/vertical-pod-autoscaler/deploy/vpa-v1-crd-gen.yaml",
		); err != nil {
			return err
		}
		if err := runCmd("Installing Vertical Pod Autoscaler RBAC", "kubectl", "apply", "-f",
			"https://raw.githubusercontent.com/kubernetes/autoscaler/master/vertical-pod-autoscaler/deploy/vpa-rbac.yaml",
		); err != nil {
			return err
		}
		if err := runCmd("Installing Vertical Pod Autosscaler Recommender", "kubectl", "apply", "-f",
			"https://raw.githubusercontent.com/kubernetes/autoscaler/master/vertical-pod-autoscaler/deploy/recommender-deployment.yaml", //nolint:lll
		); err != nil {
			return err
		}
	}

	return nil
}

type binary struct {
	name string
	link string
}

var (
	kubectl = binary{
		name: "kubectl",
		link: "https://kubernetes.io/docs/tasks/tools",
	}
	kind = binary{
		name: "kind",
		link: "https://kind.sigs.k8s.io/docs/user/quick-start/#installation",
	}
	helm = binary{
		name: "helm",
		link: "https://helm.sh/docs/intro/install/",
	}
	docker = binary{
		name: "docker",
		link: "https://docs.docker.com/engine/install/",
	}
)

func checkBinaries(binaries ...binary) error {
	hasAll := true
	for _, bin := range binaries {
		if _, err := exec.LookPath(bin.name); err != nil {
			fmt.Printf(
				"No bin bin.named '%s' could be found. Install %s and make sure it's in the PATH to use this command\n",
				bin.name,
				bin.name,
			)
			fmt.Printf("See here %s for how to install %s\n\n", bin.link, bin.name)
			hasAll = false
		}
	}

	if !hasAll {
		return errors.New("missing binaries")
	}

	return nil
}

func runCmd(displayMessage string, arg string, args ...string) error {
	var err error
	cmd := common.NewDeferredOutputCommand(displayMessage, common.Verbose(verbose))
	defer func() {
		cmd.End(err == nil)
	}()
	err = cmd.RunNew(arg, args...)
	return err
}

func (c *Cmd) runInit(_ *cobra.Command, _ []string) error {
	execCmd := exec.Command(
		"kubectl", "exec", "--tty", "--stdin",
		"--namespace", "rig-system",
		"deploy/rig-platform", "--",
		"rig-admin", "init", installationID,
	)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}
