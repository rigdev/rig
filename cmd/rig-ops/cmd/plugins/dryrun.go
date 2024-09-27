package plugins

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Cmd) dryRun(ctx context.Context, _ *cobra.Command, args []string) error {
	cfg, err := base.GetOperatorConfig(ctx, c.OperatorClient, c.Scheme)
	if err != nil {
		return err
	}

	for _, r := range replaces {
		idx, path, err := parseReplace(r)
		if err != nil {
			return fmt.Errorf("replace '%s' was malfored: %q", r, err)
		}
		if idx >= len(cfg.Pipeline.Steps) {
			return fmt.Errorf("replace idx %v too high (only %v steps)", idx, len(cfg.Pipeline.Steps))
		}
		step, err := readPlugin(path)
		if err != nil {
			return err
		}
		cfg.Pipeline.Steps[idx] = step
	}

	idx := 0
	for i, step := range cfg.Pipeline.Steps {
		if !slices.Contains(removes, i+1) {
			cfg.Pipeline.Steps[idx] = step
			idx++
		}
	}
	cfg.Pipeline.Steps = cfg.Pipeline.Steps[:idx]

	for _, a := range appends {
		step, err := readPlugin(a)
		if err != nil {
			return err
		}
		cfg.Pipeline.Steps = append(cfg.Pipeline.Steps, step)
	}

	if dry {
		o := common.OutputTypeYAML
		if base.Flags.OutputType != common.OutputTypePretty {
			o = base.Flags.OutputType
		}
		return common.FormatPrint(cfg.Pipeline.Steps, o)
	}

	if specPath != "" && len(args) > 0 {
		return fmt.Errorf("can't supply both --spec and capsule name")
	}

	var spec string
	var capsuleSpec v1alpha2.Capsule
	if len(args) > 0 {
		if err := c.K8s.Get(ctx, client.ObjectKey{
			Namespace: args[0],
			Name:      args[1],
		}, &capsuleSpec); err != nil {
			return err
		}
	} else if specPath != "" {
		bytes, err := os.ReadFile(specPath)
		if err != nil {
			return err
		}
		spec = string(bytes)
		if err := obj.Decode([]byte(spec), &capsuleSpec); err != nil {
			return err
		}
	} else {
		capsuleList := v1alpha2.CapsuleList{}
		if err := c.K8s.List(ctx, &capsuleList); err != nil {
			return err
		}
		var choices [][]string
		for _, c := range capsuleList.Items {
			choices = append(choices, []string{c.Namespace, c.Name})
		}
		idx, err := c.Prompter.TableSelect(
			"Choose a capsule", choices, []string{"Namespace", "Capsule"}, common.SelectEnableFilterOpt,
		)
		if err != nil {
			return err
		}
		choice := choices[idx]
		if err := c.K8s.Get(ctx, client.ObjectKey{
			Namespace: choice[0],
			Name:      choice[1],
		}, &capsuleSpec); err != nil {
			return err
		}
	}

	cfgBytes, err := obj.EncodeAny(cfg)
	if err != nil {
		return err
	}

	dryRun, err := c.OperatorClient.Pipeline.DryRun(ctx, connect.NewRequest(&pipeline.DryRunRequest{
		Namespace:      capsuleSpec.Namespace,
		Capsule:        capsuleSpec.Name,
		OperatorConfig: string(cfgBytes),
		CapsuleSpec:    spec,
	}))
	if err != nil {
		return err
	}

	dryOutput, err := c.processDryRunOutput(dryRun.Msg)
	if err != nil {
		return err
	}

	if interactive {
		return capsule.PromptDryOutput(ctx, dryOutput, c.Scheme)
	}

	var objects []any
	for _, o := range dryOutput.KubernetesObjects {
		if o.New.Object != nil {
			objects = append(objects, o.New.Object)
		}
	}

	out, err := common.Format(objects, common.OutputTypeYAML)
	if err != nil {
		return err
	}

	if output == "" {
		fmt.Println(out)
		return nil
	}

	return os.WriteFile(output, []byte(out), 0o666)
}

func (c *Cmd) processDryRunOutput(resp *pipeline.DryRunResponse) (capsule.DryOutput, error) {
	var res capsule.DryOutput

	objects := map[string]capsule.KubernetesDryObject{}
	var err error
	for _, o := range resp.GetInputObjects() {
		name := fmt.Sprintf("%s %s", o.GetGvk(), o.GetName())
		var co client.Object
		if content := o.GetContent(); content != "" {
			co, err = obj.DecodeUnstructured([]byte(content))
			if err != nil {
				return capsule.DryOutput{}, err
			}
		}
		k8s := objects[name]
		k8s.Old = capsule.KubernetesObject{
			Object: co,
			YAML:   o.GetContent(),
		}
		objects[name] = k8s
	}

	for _, oo := range resp.GetOutputObjects() {
		o := oo.GetObject()
		name := fmt.Sprintf("%s %s", o.GetGvk(), o.GetName())
		var co client.Object
		if content := o.GetContent(); content != "" {
			co, err = obj.DecodeUnstructured([]byte(content))
			if err != nil {
				return capsule.DryOutput{}, err
			}
		}
		k8s := objects[name]
		k8s.New = capsule.KubernetesObject{
			Object: co,
			YAML:   o.GetContent(),
		}
		objects[name] = k8s
	}

	for _, key := range slices.Sorted(maps.Keys(objects)) {
		res.KubernetesObjects = append(res.KubernetesObjects, objects[key])
	}

	return res, nil
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
	idxStr, path := replace[:idx-1], replace[idx:]
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return 0, "", err
	}
	return idx, path, nil
}
