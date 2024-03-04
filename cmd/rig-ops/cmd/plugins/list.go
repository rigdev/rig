package plugins

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
)

func list(ctx context.Context,
	_ *cobra.Command,
	_ []string,
	operatorClient *base.OperatorClient,
	rc rig.Client,
	scheme *runtime.Scheme,
) error {
	cfg, err := getOperatorConfig(ctx, operatorClient, scheme)
	if err != nil {
		return err
	}

	var plugins []pluginStep
	for _, p := range cfg.Steps {
		plugin := pluginStep{
			Plugin:     p.Plugin,
			Namespaces: p.Namespaces,
			Capsules:   p.Capsules,
		}
		if showConfig {
			plugin.Config = map[string]any{}
			if err := yaml.Unmarshal([]byte(p.Config), &plugin.Config); err != nil {
				return fmt.Errorf("plugin '%s' had malformed config: %q", p.Plugin, err)
			}
		}
		plugins = append(plugins, plugin)
	}

	if base.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(&plugins, base.Flags.OutputType)
	}

	if showConfig {
		return common.FormatPrint(&plugins, common.OutputTypeYAML)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Index", "Plugin", "Namespaces", "Capsules")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for idx, p := range plugins {
		n := max(1, len(p.Capsules), len(p.Namespaces))
		for i := 0; i < n; i++ {
			col1, col2, default_ := "", "", ""
			if i == 0 {
				col1, col2, default_ = strconv.Itoa(idx), p.Plugin, "Matches all"
			}
			tbl.AddRow(col1, col2, getString(p.Namespaces, i, default_), getString(p.Capsules, i, default_))
		}
	}
	tbl.Print()

	return nil
}

type pluginStep struct {
	Plugin     string         `json:"plugin"`
	Namespaces []string       `json:"namespaces"`
	Capsules   []string       `json:"capsules"`
	Config     map[string]any `json:"config,omitempty"`
}

func getString(strings []string, idx int, default_ string) string {
	if idx < len(strings) {
		return strings[idx]
	}
	return default_
}

func get(ctx context.Context,
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

	var idx int
	if len(args) == 0 {
		if len(cfg.Steps) == 0 {
			fmt.Println("operator has no plugins configured")
			return nil
		}
		var choices [][]string
		for idx, s := range cfg.Steps {
			choices = append(choices, []string{strconv.Itoa(idx), s.Plugin})
		}
		idx, err = common.PromptTableSelect("Choose a plugin", choices, []string{"Index", "Type"})
		if err != nil {
			return err
		}
	} else {
		idx, err = strconv.Atoi(args[0])
		if err != nil {
			return err
		}
	}

	if idx >= len(cfg.Steps) {
		return fmt.Errorf("there are %v plugins configured. Max index allowed is %v", len(cfg.Steps), len(cfg.Steps)-1)
	}

	p := cfg.Steps[idx]
	plugin := pluginStep{
		Plugin:     p.Plugin,
		Namespaces: p.Namespaces,
		Capsules:   p.Capsules,
		Config:     map[string]any{},
	}
	if err := yaml.Unmarshal([]byte(p.Config), &plugin.Config); err != nil {
		return fmt.Errorf("plugin had malformed config: %q", err)
	}

	outputType := common.OutputTypeYAML
	if base.Flags.OutputType != common.OutputTypePretty {
		outputType = base.Flags.OutputType
	}

	return common.FormatPrint(plugin, outputType)
}
