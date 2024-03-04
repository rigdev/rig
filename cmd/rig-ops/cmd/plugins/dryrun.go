package plugins

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
)

func dryRun(ctx context.Context,
	_ *cobra.Command,
	args []string,
	operatorClient *base.OperatorClient,
	rc rig.Client,
	scheme *runtime.Scheme,
) error {
	cfg, err := getOperatorConfig(ctx, operatorClient, scheme)
	if err != nil {
		return err
	}

	if pluginConfig != "" {
		bytes, err := os.ReadFile(pluginConfig)
		if err != nil {
			return err
		}
		cfg.Steps = nil // Is this necessary?
		if err := yaml.Unmarshal(bytes, &cfg.Steps); err != nil {
			return err
		}
	}

	for _, r := range replaces {
		idx, path, err := parseReplace(r)
		if err != nil {
			return fmt.Errorf("replace '%s' was malfored: %q", r, err)
		}
		if idx >= len(cfg.Steps) {
			return fmt.Errorf("replace idx %v too high (only %v steps)", idx, len(cfg.Steps))
		}
		step, err := readPlugin(path)
		if err != nil {
			return err
		}
		cfg.Steps[idx] = step
	}

	idx := 0
	for i, step := range cfg.Steps {
		if !slices.Contains(removes, i) {
			cfg.Steps[idx] = step
			idx += 1
		}
	}
	cfg.Steps = cfg.Steps[:idx]

	for _, a := range appends {
		step, err := readPlugin(a)
		if err != nil {
			return err
		}
		cfg.Steps = append(cfg.Steps, step)
	}

	if dry {
		o := common.OutputTypeYAML
		if base.Flags.OutputType != common.OutputTypePretty {
			o = base.Flags.OutputType
		}
		return common.FormatPrint(cfg.Steps, o)
	}

	if specPath != "" && len(args) > 0 {
		return fmt.Errorf("can't supply both --spec and capsule name")
	}

	var spec string
	if specPath != "" {
		bytes, err := os.ReadFile(specPath)
		if err != nil {
			return err
		}
		spec = string(bytes)
	}

	var capsule string
	if len(args) > 0 {
		capsule = args[0]
	} else {
		resp, err := rc.Capsule().List(ctx, connect.NewRequest(&capsule_api.ListRequest{
			ProjectId: base.Flags.Project,
		}))
		if err != nil {
			return err
		}
		var capsules []string
		for _, c := range resp.Msg.GetCapsules() {
			capsules = append(capsules, c.GetCapsuleId())
		}
		idx, _, err := common.PromptSelect("Choose a capsule", capsules)
		if err != nil {
			return err
		}
		capsule = capsules[idx]
	}

	resp, err := rc.Environment().GetNamespaces(ctx, connect.NewRequest(&environment.GetNamespacesRequest{
		ProjectEnvs: []*environment.ProjectEnvironment{{
			ProjectId:     base.Flags.Project,
			EnvironmentId: base.Flags.Environment,
		}},
	}))
	if err != nil {
		return err
	}
	namespace := resp.Msg.GetNamespaces()[0].GetNamespace()

	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dryRun, err := operatorClient.Pipeline.DryRun(ctx, connect.NewRequest(&pipeline.DryRunRequest{
		Namespace:      namespace,
		Capsule:        capsule,
		OperatorConfig: string(cfgBytes),
		CapsuleSpec:    spec,
	}))
	if err != nil {
		return err
	}

	var objects []any
	for _, o := range dryRun.Msg.GetOutputObjects() {
		object, err := obj.DecodeAny([]byte(o.GetObject().GetContent()), scheme)
		if err != nil {
			return err
		}
		objects = append(objects, object)
	}

	out, err := common.Format(objects, common.OutputTypeYAML)
	if err != nil {
		return err
	}

	if output == "" {
		fmt.Println(out)
		return nil
	}

	return os.WriteFile(output, []byte(out), 0666)
}

func readPlugin(path string) (v1alpha1.Step, error) {
	var step v1alpha1.Step
	bytes, err := os.ReadFile(path)
	if err != nil {
		return step, err
	}
	if err := yaml.Unmarshal(bytes, &step); err != nil {
		return step, err
	}
	return step, nil
}

func parseReplace(replace string) (int, string, error) {
	idx := strings.Index(replace, ":")
	if idx == -1 {
		return 0, "", errors.New("missing ':'")
	}
	idxStr, path := replace[:idx], replace[idx+1:]
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return 0, "", err
	}
	return idx, path, nil
}
