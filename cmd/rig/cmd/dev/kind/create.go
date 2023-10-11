package kind

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kind/pkg/cluster"
)

//go:embed config.yaml
var config string

//go:embed registry.yaml
var registry string

func (c Cmd) create(cmd *cobra.Command, args []string) error {
	if err := checkBinaries(kubectl, kind, helm); err != nil {
		return err
	}

	if err := setupKindRigCluster(); err != nil {
		return err
	}

	if err := setupK8s(); err != nil {
		return err
	}

	if err := helmInstall(); err != nil {
		return err
	}

	if err := c.deploy(cmd, args); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("To use Rig you need to create at least one admin user.")
	if err := runCmd("kubectl", "exec", "--tty", "--stdin", "--namespace", "rig-system", "deploy/rig-platform", "--", "rig-admin", "init"); err != nil {
		return err
	}

	return nil
}

func (c Cmd) deploy(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	if err := checkBinaries(kind, kubectl, helm, docker); err != nil {
		return err
	}

	if operatorDockerTag == "" {
		operatorDockerTag = "latest"
	}
	if err := c.deployInner(ctx, deployParams{
		dockerImage: "ghcr.io/rigdev/rig-operator",
		dockerTag:   operatorDockerTag,
		chartName:   "rig-operator",
		chartPath:   operatorChartPath,
		customArgs:  []string{"--set", fmt.Sprintf("image.tag=%s", operatorDockerTag)},
	}); err != nil {
		return err
	}

	if platformDockerTag == "" {
		platformDockerTag = "latest"
	}
	if err := c.deployInner(ctx, deployParams{
		dockerImage: "ghcr.io/rigdev/rig-platform",
		dockerTag:   platformDockerTag,
		chartName:   "rig-platform",
		chartPath:   platformChartPath,
		customArgs: []string{
			"--set", fmt.Sprintf("image.tag=%s", platformDockerTag),
			"--set", "rig.telemetry.enabled=false",
			"--set", "postgres.enabled=true",
			"--set", "rig.cluster.dev_registry.host=localhost:30000",
			"--set", "rig.cluster.dev_registry.cluster_host=registry:5000",
			"--set", "loadBalancer.enabled=true",
		},
	}); err != nil {
		return err
	}
	fmt.Println()

	return nil
}

func waitUntilDeploymentIsReady(deployment string, humanReadableName string) error {
	fmt.Printf("Waiting for %s to be ready....\n", humanReadableName)
	type ready struct {
		Status struct {
			Replicas            int `yaml:"replicas,omitempty"`
			UnavailableReplicas int `yaml:"unavailableReplicas,omitempty"`
			AvailableReplicas   int `yaml:"availableReplicas,omitempty"`
			UpdatedReplicas     int `yaml:"updatedReplicas,omitempty"`
		} `yaml:"status,omitempty"`
	}
	c := 0
	for {
		out, err := exec.Command("kubectl", "--context", "kind-rig", "get", deployment, "-n", "rig-system", "-oyaml").Output()
		if err != nil {
			c++
			if c > 20 {
				return err
			}

			time.Sleep(500 * time.Millisecond)
			continue
		}

		var r ready
		if err := yaml.Unmarshal(out, &r); err != nil {
			return err
		}
		fmt.Printf("%+v\n", r)
		if r.Status.Replicas >= 1 &&
			r.Status.AvailableReplicas == r.Status.Replicas &&
			r.Status.UpdatedReplicas == r.Status.Replicas {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	fmt.Printf("%s is ready!\n", humanReadableName)
	return nil
}

type deployParams struct {
	dockerImage string
	dockerTag   string
	chartName   string
	chartPath   string
	customArgs  []string
}

func (c Cmd) deployInner(ctx context.Context, p deployParams) error {
	if err := c.loadImage(ctx, p.dockerImage, p.dockerTag); err != nil {
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
		"--set", fmt.Sprintf("image.tag=%s", operatorDockerTag),
		"--create-namespace",
	}
	cArgs = append(cArgs, p.customArgs...)
	if p.chartPath == "" {
		cArgs = append(cArgs, "--repo", "https://charts.rig.dev")
	}
	if err := runCmd("helm", cArgs...); err != nil {
		return err
	}

	if err := runCmd("kubectl", "--context", "kind-rig", "rollout", "restart", "deployment", "-n", "rig-system", p.chartName); err != nil {
		return err
	}

	if err := waitUntilDeploymentIsReady(fmt.Sprintf("deployment.apps/%s", p.chartName), p.chartName); err != nil {
		return err
	}

	return nil
}

func (c Cmd) loadImage(ctx context.Context, image, tag string) error {
	imageTag := fmt.Sprintf("%s:%s", image, tag)
	res, err := c.DockerClient.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: imageTag,
		}),
	})
	if err != nil {
		return err
	}
	if len(res) == 0 || tag == "latest" {
		if err := runCmd("docker", "pull", imageTag); err != nil {
			return err
		}
	}

	if err := runCmd("kind", "load", "docker-image", imageTag, "-n", "rig"); err != nil {
		return err
	}

	return nil
}

func (c Cmd) clean(cmd *cobra.Command, args []string) error {
	if err := checkBinaries(kind); err != nil {
		return err
	}

	if err := runCmd("kind", "delete", "clusters", "rig"); err != nil {
		return err
	}

	return nil
}

func setupKindRigCluster() error {
	provider := cluster.NewProvider()
	clusters, err := provider.List()
	if err != nil {
		return err
	}

	if slices.Contains(clusters, "rig") {
		return nil
	}

	if err := provider.Create(
		"rig",
		cluster.CreateWithRawConfig([]byte(config)),
	); err != nil {
		return err
	}

	return nil
}

func setupK8s() error {
	if err := runCmd("kubectl", "--context", "kind-rig", "get", "namespace", "rig-system"); err != nil {
		if err := runCmd("kubectl", "--context", "kind-rig", "create", "namespace", "rig-system"); err != nil {
			return err
		}
	}

	cmd := exec.Command("kubectl", "--context", "kind-rig", "apply", "-n", "rig-system", "-f", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close() // the doc says subProcess.Wait will close it, but I'm not sure, so I kept this line

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Start(); err != nil {
		return err
	}

	io.WriteString(stdin, registry)
	stdin.Close()
	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func helmInstall() error {
	if err := runCmd(
		"helm", "--kube-context", "kind-rig",
		"upgrade", "--install", "cert-manager", "cert-manager",
		"--repo", "https://charts.jetstack.io",
		"--namespace", "cert-manager",
		"--create-namespace", "--version", "v1.13.0",
		"--set", "installCRDs=true",
	); err != nil {
		return err
	}

	if err := runCmd(
		"helm", "--kube-context", "kind-rig",
		"upgrade", "--install", "metrics-server", "metrics-server",
		"--repo", "https://kubernetes-sigs.github.io/metrics-server",
		"--namespace", "kube-system",
		"--set", "args={--kubelet-insecure-tls}",
	); err != nil {
		return err
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
			fmt.Printf("No bin bin.named '%s' could be found. Install %s and make sure it's in the PATH to use this command\n", bin.name, bin.name)
			fmt.Printf("See here %s for how to install %s\n\n", bin.link, bin.name)
			hasAll = false
		}
	}

	if !hasAll {
		return errors.New("missing binaries")
	}

	return nil
}

func runCmd(arg string, args ...string) error {
	cmd := exec.Command(arg, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
