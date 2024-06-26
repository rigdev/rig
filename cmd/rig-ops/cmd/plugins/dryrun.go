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
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/migrate"
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

	if pluginConfig != "" {
		bytes, err := os.ReadFile(pluginConfig)
		if err != nil {
			return err
		}
		cfg.Pipeline.Steps = nil // Is this necessary?
		if err := yaml.Unmarshal(bytes, &cfg.Pipeline); err != nil {
			return err
		}
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
		if !slices.Contains(removes, i) {
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
	var capsule v1alpha2.Capsule
	if len(args) > 0 {
		if err := c.K8s.Get(ctx, client.ObjectKey{
			Namespace: args[0],
			Name:      args[1],
		}, &capsule); err != nil {
			return err
		}
	} else if specPath != "" {
		bytes, err := os.ReadFile(specPath)
		if err != nil {
			return err
		}
		spec = string(bytes)
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
		}, &capsule); err != nil {
			return err
		}
	}

	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dryRun, err := c.OperatorClient.Pipeline.DryRun(ctx, connect.NewRequest(&pipeline.DryRunRequest{
		Namespace:      capsule.Namespace,
		Capsule:        capsule.Name,
		OperatorConfig: string(cfgBytes),
		CapsuleSpec:    spec,
	}))
	if err != nil {
		return err
	}

	var objects []any
	for _, o := range dryRun.Msg.GetOutputObjects() {
		object, err := obj.DecodeAny([]byte(o.GetObject().GetContent()), c.Scheme)
		if err != nil {
			return err
		}
		objects = append(objects, object)
	}

	if interactive {
		return c.interactiveDiff(dryRun.Msg)
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

func (c *Cmd) interactiveDiff(dryRun *pipeline.DryRunResponse) error {
	current := migrate.NewResources()
	for _, o := range dryRun.InputObjects {
		object, err := obj.DecodeAny([]byte(o.GetContent()), c.Scheme)
		if err != nil {
			return err
		}
		if err := current.AddObject(o.GetGvk().Kind, o.GetName(), object); err != nil {
			return err
		}
	}
	overview := current.CreateOverview("Current Resources")

	migrated := migrate.NewResources()
	if err := migrate.ProcessOperatorOutput(migrated, dryRun.GetOutputObjects(), c.Scheme); err != nil {
		return err
	}

	migratedOverview := migrated.CreateOverview("New Resources")

	reports, err := migrated.Compare(current, c.Scheme)
	if err != nil {
		return err
	}

	warnings := map[string][]*migrate.Warning{}
	for _, k := range reports.GetKinds() {
		warnings[k] = nil
	}

	return migrate.PromptDiffingChanges(reports, warnings, overview, migratedOverview, c.Prompter)
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
