package kind

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"sigs.k8s.io/kind/pkg/cluster"
)

// embed cannot read a file not in the directory of this source file, so we need to copy it there first
//
//go:generate cp ../../../../../deploy/kind/config.yaml config.yaml
//go:embed config.yaml
var config string

//go:generate cp ../../../../../deploy/registry/registry.yaml registry.yaml
//go:embed registry.yaml
var registry string

func (c Cmd) create(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
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

	if err := c.setupKindContext(ctx); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Creating a new Rig cluster requries setting up an admin user and project")
	fmt.Println("Run the following command once the new Rig server has finished starting up")
	fmt.Println("kubectl", "exec", "--tty", "--stdin", "--namespace", "rig-system", "deploy/rig", "--", "rig-admin", "init")

	return nil
}

func (c Cmd) deploy(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	if err := checkBinaries(kind, kubectl, helm, docker); err != nil {
		return err
	}

	var err error
	if dockerTag == "" {
		dockerTag, err = getLatestTag("rig")
		if err != nil {
			return err
		}
	}
	dockerTag = strings.TrimPrefix(dockerTag, "v")
	rigImage := fmt.Sprintf("ghcr.io/rigdev/rig:%s", dockerTag)
	res, err := c.DockerClient.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: rigImage,
		}),
	})
	if err != nil {
		return err
	}
	if len(res) == 0 {
		if err := runCmd("docker", "pull", rigImage); err != nil {
			return err
		}
	}

	if err := runCmd("kind", "load", "docker-image", rigImage, "-n", "rig"); err != nil {
		return err
	}

	if helmChartTag == "" {
		helmChartTag, err = getLatestTag("charts")
		if err != nil {
			return err
		}
	}

	chart := "rig"
	if chartPath != "" {
		chart = chartPath
	}
	cArgs := []string{
		"--kube-context", "kind-rig",
		"upgrade", "--install", "rig", chart,
		"--namespace", "rig-system",
		"--version", dockerTag,
		"--set", fmt.Sprintf("image.tag=%s", dockerTag),
		"--set", "mongodb.enabled=true",
		"--set", "rig.telemetry.enabled=false",
		"--set", "rig.cluster.dev_registry.host=localhost:30000",
		"--set", "rig.cluster.dev_registry.cluster_host=registry:5000",
		"--set", "service.type=NodePort",
		"--create-namespace",
	}
	if chartPath == "" {
		cArgs = append(args, "--repo", "https://charts.rig.dev")
	}

	if err := runCmd(
		"helm", cArgs...,
	); err != nil {
		return err
	}

	if err := runCmd("kubectl", "--context", "kind-rig", "rollout", "restart", "deployment", "-n", "rig-system", "rig"); err != nil {
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

func getLatestTag(repo string) (string, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://api.github.com/repos/rigdev/%s/releases/latest", repo), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	b := make([]byte, 1000)
	n := 0
	for !errors.Is(err, io.EOF) {
		n, err = resp.Body.Read(b)
		buffer.Write(b[:n])
	}

	tag := struct {
		TagName string `json:"tag_name"`
	}{}
	if err := json.Unmarshal(buffer.Bytes(), &tag); err != nil {
		return "", err
	}

	return tag.TagName, nil
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
	output, err := exec.Command("helm", "--kube-context", "kind-rig", "repo", "list").Output()
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`\njetstack\s*https://charts.jetstack.io\s*\n`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) == 0 {
		if err := exec.Command("helm", "--kube-context", "kind-rig", "repo", "add", "jetstack", "https://charts.jetstack.io").Run(); err != nil {
			return err
		}
	}

	re = regexp.MustCompile(`\nmetrics-server\s*https://kubernetes-sigs.github.io/metrics-server\s*\n`)
	matches = re.FindStringSubmatch(string(output))
	if len(matches) == 0 {
		if err := exec.Command("helm", "--kube-context", "kind-rig", "repo", "add", "metrics-server", "https://kubernetes-sigs.github.io/metrics-server").Run(); err != nil {
			return err
		}
	}

	if err := runCmd("helm", "--kube-context", "kind-rig", "repo", "update"); err != nil {
		return err
	}

	if err := runCmd(
		"helm", "--kube-context", "kind-rig",
		"upgrade", "--install", "cert-manager", "jetstack/cert-manager",
		"--namespace", "cert-manager",
		"--create-namespace", "--version", "v1.13.0",
		"--set", "installCRDs=true",
	); err != nil {
		return err
	}

	if err := runCmd(
		"helm", "--kube-context", "kind-rig",
		"upgrade", "--install", "metrics-server", "metrics-server/metrics-server",
		"--namespace", "kube-system",
		"--set", "args={--kubelet-insecure-tls}",
	); err != nil {
		return err
	}

	return nil
}

func (c Cmd) setupKindContext(ctx context.Context) error {
	fmt.Println("")
	fmt.Println("Rig on the Kind cluster listens on the port 30047. This requires you to use a different context than the default one which uses port 4747.")
	ok, err := common.PromptConfirm("Create a new context which uses port 30047?", true)
	if err != nil {
		return err
	}
	if ok {
		if err := cmd_config.CreateContext(c.Cfg, "kind", "http://localhost:30047/"); err != nil {
			return err
		}
		return nil
	}

	ok, err = common.PromptConfirm("Change context?", true)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return cmd_config.SelectContext(c.Cfg)
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
