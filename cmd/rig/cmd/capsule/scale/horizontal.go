package scale

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/validation"
)

func (c *Cmd) horizontal(ctx context.Context, cmd *cobra.Command, _ []string) error {
	horizontal := &capsule.HorizontalScale{}

	rollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil && !errors.IsNotFound(err) {
		return nil
	}

	if rollout.GetConfig() != nil {
		horizontal = rollout.GetConfig().GetHorizontalScale()
	}

	if horizontal.CpuTarget != nil && !overwriteAutoscaler {
		return errors.New(
			"cannot set the number of replicas with the autoscaler enabled with setting the --overwrite-autoscaler flag",
		)
	}

	horizontal.CpuTarget = nil

	if !cmd.Flags().Lookup("replicas").Changed {
		return errors.New("--replicas not set")
	}
	horizontal.MinReplicas = replicas
	horizontal.MaxReplicas = replicas

	req := connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Changes: []*capsule.Change{
			{
				Field: &capsule.Change_HorizontalScale{
					HorizontalScale: horizontal,
				},
			},
		},
		ProjectId:     flags.GetProject(c.Cfg),
		EnvironmentId: flags.GetEnvironment(c.Cfg),
	})

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, c.Cfg, capsule_cmd.CapsuleID, req)
		} else {
			_, err = capsule_cmd.PromptAbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, c.Cfg, req)
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *Cmd) autoscale(ctx context.Context, cmd *cobra.Command, _ []string) error {
	var replicas uint32
	horizontal := &capsule.HorizontalScale{}

	rollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if rollout.GetConfig() != nil {
		replicas = rollout.GetConfig().GetReplicas()
		horizontal = rollout.GetConfig().GetHorizontalScale()
	}

	if autoscalerPath != "" {
		bytes, err := os.ReadFile(autoscalerPath)
		if err != nil {
			return err
		}

		var raw interface{}
		if err := yaml.Unmarshal(bytes, &raw); err != nil {
			return err
		}

		if bytes, err = json.Marshal(raw); err != nil {
			return err
		}

		if err := protojson.Unmarshal(bytes, horizontal); err != nil {
			return err
		}
	}

	if cmd.Flags().Lookup("min-replicas").Changed {
		horizontal.MinReplicas = uint32(minReplicas)
	}
	if cmd.Flags().Lookup("max-replicas").Changed {
		horizontal.MaxReplicas = uint32(maxReplicas)
	}
	if cmd.Flags().Lookup("utilization-percentage").Changed {
		cpuTarget := horizontal.GetCpuTarget()
		if cpuTarget == nil {
			cpuTarget = &capsule.CPUTarget{}
		}
		cpuTarget.AverageUtilizationPercentage = uint32(utilizationPercentage)
		horizontal.CpuTarget = cpuTarget
	}

	if disable {
		horizontal.CpuTarget = nil
		horizontal.MinReplicas = replicas
		horizontal.MaxReplicas = replicas
		horizontal.CustomMetrics = nil
	}

	if !hasAutoscalerFlagsSet(cmd) {
		if err := c.promptAutoscale(ctx, horizontal); err != nil {
			return err
		}
	}

	req := connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Changes: []*capsule.Change{
			{
				Field: &capsule.Change_HorizontalScale{
					HorizontalScale: horizontal,
				},
			},
		},
		ProjectId:     flags.GetProject(c.Cfg),
		EnvironmentId: flags.GetEnvironment(c.Cfg),
	})

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, c.Cfg, capsule_cmd.CapsuleID, req)
		} else {
			_, err = capsule_cmd.PromptAbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, c.Cfg, req)
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func hasAutoscalerFlagsSet(cmd *cobra.Command) bool {
	return (autoscalerPath != "" ||
		cmd.Flags().Lookup("min-replicas").Changed ||
		cmd.Flags().Lookup("max-replicas").Changed ||
		cmd.Flags().Lookup("utilization-percentage").Changed ||
		disable)
}

func (c *Cmd) promptAutoscale(ctx context.Context, horizontal *capsule.HorizontalScale) error {
	for {
		idx, _, err := common.PromptSelect("Choose action", []string{
			"See configuration",
			"Save and finish",
			"Set min instances",
			"Set max instances",
			"Set CPU utilization percentage",
			"Add custom metric",
			"Remove custom metric",
		})
		if err != nil {
			return err
		}
		switch idx {
		case 0:
			// TODO Fix this hack!
			o := flags.Flags.OutputType
			flags.Flags.OutputType = common.OutputTypeYAML
			if err := common.FormatPrint(horizontal, flags.Flags.OutputType); err != nil {
				return err
			}
			flags.Flags.OutputType = o
		case 1:
			if err := validateAutoscaler(horizontal); err != nil {
				fmt.Println(err)
				continue
			}
			return nil
		case 2:
			s, err := common.PromptInput("Min instances:", common.ValidateIntOpt)
			if err != nil {
				return err
			}
			m, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			horizontal.MinReplicas = uint32(m)
		case 3:
			s, err := common.PromptInput("Max instances:", common.ValidateIntOpt)
			if err != nil {
				return err
			}
			m, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			horizontal.MaxReplicas = uint32(m)
		case 4:
			s, err := common.PromptInput("Utilization percentage:", common.ValidateIntInRangeOpt(0, 100))
			if err != nil {
				return err
			}
			p, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			if horizontal.GetCpuTarget() == nil {
				horizontal.CpuTarget = &capsule.CPUTarget{}
			}
			horizontal.CpuTarget.AverageUtilizationPercentage = uint32(p)
		case 5:
			metric, err := c.promptCustomMetric(ctx)
			if err != nil {
				return err
			}
			horizontal.CustomMetrics = append(horizontal.CustomMetrics, metric)
		case 6:
			if len(horizontal.GetCustomMetrics()) == 0 {
				fmt.Println("Configuration has no custom metrics")
				break
			}

			choices := [][]string{{
				"Go back", "-",
			}}
			for _, m := range horizontal.GetCustomMetrics() {
				var kind string
				var name string
				if o := m.GetObject(); o != nil {
					kind = "Object"
					name = o.GetMetricName()
				} else if i := m.GetInstance(); i != nil {
					kind = "Instance"
					name = i.GetMetricName()
				}
				choices = append(choices, []string{kind, name})
			}
			idx, err := common.PromptTableSelect("Select Metric", choices, []string{"Type", "Name"})
			if err != nil {
				return err
			}
			if idx == 0 {
				break
			}
			horizontal.CustomMetrics = append(horizontal.CustomMetrics[:idx-1], horizontal.CustomMetrics[idx:]...)
		}
	}
}

func validateAutoscaler(horizontal *capsule.HorizontalScale) error {
	if horizontal.MaxReplicas < horizontal.MinReplicas {
		return errors.New("max-replicas cannot be smaller than min-replicas")
	}

	return nil
}

func (c *Cmd) promptCustomMetric(ctx context.Context) (*capsule.CustomMetric, error) {
	idx, _, err := common.PromptSelect("Metric type:", []string{
		"Instance Metric",
		"Object Metric",
	})
	if err != nil {
		return nil, err
	}

	switch idx {
	case 0:
		metric, err := c.promptInstanceMetric(ctx)
		if err != nil {
			return nil, err
		}
		return &capsule.CustomMetric{
			Metric: metric,
		}, nil
	case 1:
		metric, err := c.promptObjectMetric(ctx)
		if err != nil {
			return nil, err
		}
		return &capsule.CustomMetric{
			Metric: metric,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected index %v", idx)
	}
}

func (c *Cmd) promptInstanceMetric(ctx context.Context) (*capsule.CustomMetric_Instance, error) {
	metrics, err := c.Rig.Capsule().
		GetCustomInstanceMetrics(ctx, connect.NewRequest(&capsule.GetCustomInstanceMetricsRequest{
			CapsuleId:     capsule_cmd.CapsuleID,
			ProjectId:     flags.GetProject(c.Cfg),
			EnvironmentId: flags.GetEnvironment(c.Cfg),
		}))
	if err != nil {
		return nil, err
	}

	metricName, err := prompMetricName(metrics.Msg.GetMetrics())
	if err != nil {
		return nil, err
	}

	labelSelectors, err := promptLabelSelector()
	if err != nil {
		return nil, err
	}

	value, err := common.PromptInput("Average Value:", common.ValidateQuantityOpt)
	if err != nil {
		return nil, err
	}

	return &capsule.CustomMetric_Instance{
		Instance: &capsule.InstanceMetric{
			MetricName:   metricName,
			MatchLabels:  labelSelectors,
			AverageValue: value,
		},
	}, nil
}

func (c *Cmd) promptObjectMetric(ctx context.Context) (*capsule.CustomMetric_Object, error) {
	kind, err := common.PromptInput("Described object, kind:", common.ValidateKubernetesNameOpt)
	if err != nil {
		return nil, err
	}

	resp, err := c.Rig.Project().GetObjectsByKind(ctx, connect.NewRequest(&project.GetObjectsByKindRequest{
		Kind:          kind,
		ProjectId:     flags.GetProject(c.Cfg),
		EnvironmentId: flags.GetEnvironment(c.Cfg),
	}))

	var objName string
	if err != nil {
		objName, err = common.PromptInput("Object name:")
		if err != nil {
			return nil, err
		}
	} else {
		var names []string
		for _, obj := range resp.Msg.GetObjects() {
			names = append(names, obj.GetName())
		}
		_, objName, err = common.PromptSelect("Object by name:", names, common.SelectEnableFilterOpt)
		if err != nil {
			return nil, err
		}
	}

	api, err := common.PromptInput("Described object, api version (optional):", func(inp *textinput.TextInput) {
		inp.Validate = func(s string) error {
			if s == "" {
				return nil
			}
			return common.ValidateKubernetesName(s)
		}
	},
	)
	if err != nil {
		return nil, err
	}

	metricResp, err := c.Rig.Project().GetCustomObjectMetrics(
		ctx,
		connect.NewRequest(&project.GetCustomObjectMetricsRequest{
			ObjectReference: &capsule.ObjectReference{
				Kind:       kind,
				Name:       objName,
				ApiVersion: api,
			},
			ProjectId:     flags.GetProject(c.Cfg),
			EnvironmentId: flags.GetEnvironment(c.Cfg),
		}))
	if err != nil {
		return nil, err
	}

	metricName, err := prompMetricName(metricResp.Msg.GetMetrics())
	if err != nil {
		return nil, err
	}

	labelSelectors, err := promptLabelSelector()
	if err != nil {
		return nil, err
	}

	idx, s, err := common.PromptSelect("Type:", []string{"Value", "Average Value"})
	if err != nil {
		return nil, err
	}

	value, err := common.PromptInput(s+":", common.ValidateQuantityOpt)
	if err != nil {
		return nil, err
	}

	metric := &capsule.CustomMetric_Object{
		Object: &capsule.ObjectMetric{
			MetricName:  metricName,
			MatchLabels: labelSelectors,
			ObjectReference: &capsule.ObjectReference{
				Kind:       kind,
				Name:       objName,
				ApiVersion: api,
			},
		},
	}

	if idx == 0 {
		metric.Object.Value = value
	} else {
		metric.Object.AverageValue = value
	}

	return metric, nil
}

func prompMetricName(metrics []*capsule.Metric) (string, error) {
	slices.SortFunc(metrics, func(m1, m2 *capsule.Metric) int {
		return strings.Compare(m1.Name, m2.Name)
	})

	var choices [][]string
	now := time.Now()
	for _, m := range metrics {
		choices = append(choices, []string{
			m.Name,
			fmt.Sprintf("%.2f", m.GetLatestValue()),
			common.FormatDuration(now.Sub(m.GetLatestTimestamp().AsTime())),
		})
	}

	if len(choices) == 0 {
		return common.PromptInput("Metric Name:", common.ValidateKubernetesNameOpt)
	}
	idx, err := common.PromptTableSelect(
		"Select metric:",
		choices,
		[]string{"Metric", "Latest value", "Age of latest value"},
		common.SelectEnableFilterOpt,
	)
	if err != nil {
		return "", err
	}
	return metrics[idx].Name, nil
}

func promptLabelSelector() (map[string]string, error) {
	s, err := common.PromptInput("Label Selectors", func(inp *textinput.TextInput) {
		inp.Validate = func(s string) error {
			_, err := parseLabelSelectors(s)
			return err
		}
	})
	if err != nil {
		return nil, err
	}

	return parseLabelSelectors(s)
}

var errLabelSelector = errors.New("must be a space-separated list of key/value pairs of the form 'k=v'")

func parseLabelSelectors(s string) (map[string]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	splits := strings.Split(s, " ")
	idx := 0
	for _, s := range splits {
		s = strings.TrimSpace(s)
		if s != "" {
			splits[idx] = s
			idx++
		}
	}
	splits = splits[:idx]

	result := map[string]string{}
	for _, s := range splits {
		ss := strings.Split(s, "=")
		if len(ss) != 2 {
			return nil, errLabelSelector
		}

		key, value := ss[0], ss[1]
		if errs := validation.IsQualifiedName(key); errs != nil {
			return nil, fmt.Errorf("key is invalid: %s", strings.Join(errs, "; "))
		}
		if errs := validation.IsValidLabelValue(value); errs != nil {
			return nil, fmt.Errorf("value is invalid: %s", strings.Join(errs, "; "))
		}
		result[key] = value
	}

	return result, nil
}
