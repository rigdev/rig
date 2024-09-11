package plugins

import (
	"context"
	"os"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Cmd) computeConfig(ctx context.Context, _ *cobra.Command, args []string) error {
	cfg, err := base.GetOperatorConfig(ctx, c.OperatorClient, c.Scheme)
	if err != nil {
		return err
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
		if err := obj.Decode([]byte(spec), &capsule); err != nil {
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
		}, &capsule); err != nil {
			return err
		}
	}

	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	resp, err := c.OperatorClient.Pipeline.DryRunPluginConfig(ctx, connect.NewRequest(&pipeline.DryRunPluginConfigRequest{
		Namespace:      capsule.Namespace,
		Capsule:        capsule.Name,
		OperatorConfig: string(cfgBytes),
		CapsuleSpec:    spec,
	}))
	if err != nil {
		return err
	}

	return common.FormatPrint(resp.Msg.GetSteps(), common.OutputTypeYAML)
}
