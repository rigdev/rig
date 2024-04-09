package plugins

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func (c *Cmd) listSteps(ctx context.Context, _ *cobra.Command, _ []string) error {
	cfg, err := base.GetOperatorConfig(ctx, c.OperatorClient, c.Scheme)
	if err != nil {
		return err
	}
	var plugins []pluginStep
	for _, s := range cfg.Pipeline.Steps {
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
			step.Plugins = append(step.Plugins, plugin)
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
	tbl := table.New("StepIndex", "Plugins", "Namespaces", "Capsules")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for idx, p := range plugins {
		var pluginNames []string
		for _, pp := range p.Plugins {
			pluginNames = append(pluginNames, pp.Name)
		}

		n := max(1, len(p.Capsules), len(p.Namespaces), len(p.Plugins))
		for i := 0; i < n; i++ {
			col1, def := "", ""
			if i == 0 {
				col1, def = strconv.Itoa(idx), "Matches all"
			}
			tbl.AddRow(col1,
				getString(pluginNames, i, ""),
				getString(p.Namespaces, i, def),
				getString(p.Capsules, i, def),
			)
		}
	}
	tbl.Print()

	return nil
}

type pluginStep struct {
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

func (c *Cmd) get(ctx context.Context, _ *cobra.Command, args []string) error {
	cfg, err := base.GetOperatorConfig(ctx, c.OperatorClient, c.Scheme)
	if err != nil {
		return err
	}

	var idx int
	if len(args) == 0 {
		if len(cfg.Pipeline.Steps) == 0 {
			fmt.Println("operator has no plugins configured")
			return nil
		}
		var choices [][]string
		for idx, s := range cfg.Pipeline.Steps {
			var plugins []string
			for _, p := range s.Plugins {
				plugins = append(plugins, p.Name)
			}
			choices = append(choices, []string{strconv.Itoa(idx), strings.Join(plugins, ", ")})
		}
		idx, err = c.Prompter.TableSelect("Choose a plugin", choices, []string{"Index", "Plugins"})
		if err != nil {
			return err
		}
	} else {
		idx, err = strconv.Atoi(args[0])
		if err != nil {
			return err
		}
	}

	if idx >= len(cfg.Pipeline.Steps) {
		return fmt.Errorf(
			"there are %v plugins configured. Max index allowed is %v",
			len(cfg.Pipeline.Steps), len(cfg.Pipeline.Steps)-1,
		)
	}

	s := cfg.Pipeline.Steps[idx]
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

func (c *Cmd) list(ctx context.Context, _ *cobra.Command, _ []string) error {
	resp, err := c.OperatorClient.Capabilities.GetPlugins(ctx, connect.NewRequest(&capabilities.GetPluginsRequest{}))
	if err != nil {
		return err
	}

	result := struct {
		Builtin    []string     `json:"builtin,omitempty"`
		Thirdparty []thirdparty `json:"thirdparty,omitempty"`
	}{}

	for _, p := range resp.Msg.GetPlugins() {
		if b := p.GetBuiltin(); b != nil {
			result.Builtin = append(result.Builtin, b.GetName())
		} else if t := p.GetThirdParty(); t != nil {
			result.Thirdparty = append(result.Thirdparty, thirdparty{
				Name:  t.GetName(),
				Image: t.GetImage(),
			})
		}
	}

	if base.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(result, base.Flags.OutputType)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Type", "Name")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, p := range result.Builtin {
		tbl.AddRow("Builtin", p, "")
	}
	for _, p := range result.Thirdparty {
		tbl.AddRow("Thirdparty", p.Name)
	}
	tbl.Print()

	return nil
}

type thirdparty struct {
	Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
}
