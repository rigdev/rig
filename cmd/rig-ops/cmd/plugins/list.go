package plugins

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
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
	scheme *runtime.Scheme,
) error {
	cfg, err := getOperatorConfig(ctx, operatorClient, scheme)
	if err != nil {
		return err
	}

	var plugins []pluginStep
	for _, s := range cfg.Steps {
		step := pluginStep{
			Namespaces: s.Namespaces,
			Capsules:   s.Capsules,
		}
		for _, p := range s.Plugins {
			plugin := pluginInfo{
				Name: p.Name,
			}
			if showConfig {
				plugin.Config = map[string]any{}
				if err := yaml.Unmarshal([]byte(p.Config), &plugin.Config); err != nil {
					return fmt.Errorf("plugin '%s' had malformed config: %q", p.Name, err)
				}
			}
		}
		plugins = append(plugins, step)
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
			col1, col2, def := "", "", ""
			if i == 0 {
				col1, col2, def = strconv.Itoa(idx), p.Plugin, "Matches all"
			}
			tbl.AddRow(col1, col2, getString(p.Namespaces, i, def), getString(p.Capsules, i, def))
		}
	}
	tbl.Print()

	return nil
}

type pluginStep struct {
	Plugin     string       `json:"plugin"`
	Namespaces []string     `json:"namespaces"`
	Capsules   []string     `json:"capsules"`
	Plugins    []pluginInfo `json:"plugins"`
}

type pluginInfo struct {
	Name   string         `json:"name"`
	Config map[string]any `json:"config,omitempty"`
}

func getString(strings []string, idx int, def string) string {
	if idx < len(strings) {
		return strings[idx]
	}
	return def
}

func get(ctx context.Context,
	_ *cobra.Command,
	args []string,
	operatorClient *base.OperatorClient,
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
			var plugins []string
			for _, p := range s.Plugins {
				plugins = append(plugins, p.Name)
			}
			choices = append(choices, []string{strconv.Itoa(idx), strings.Join(plugins, ", ")})
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

	s := cfg.Steps[idx]
	step := pluginStep{
		Namespaces: s.Namespaces,
		Capsules:   s.Capsules,
	}
	for _, p := range s.Plugins {
		plugin := pluginInfo{
			Name: p.Name,
		}
		if err := yaml.Unmarshal([]byte(p.Config), &plugin.Config); err != nil {
			return fmt.Errorf("plugin had malformed config: %q", err)
		}
		step.Plugins = append(step.Plugins, plugin)
	}

	outputType := common.OutputTypeYAML
	if base.Flags.OutputType != common.OutputTypePretty {
		outputType = base.Flags.OutputType
	}

	return common.FormatPrint(step, outputType)
}
