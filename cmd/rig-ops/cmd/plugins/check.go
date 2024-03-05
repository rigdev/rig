package plugins

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

func check(ctx context.Context,
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
	matchers := map[string]plugin.Matcher{}
	for _, step := range cfg.Steps {
		if len(plugins) > 0 && !slices.Contains(plugins, step.Plugin) {
			continue
		}
		matcher, err := plugin.NewMatcher(step.Namespaces, step.Capsules, step.Selector)
		if err != nil {
			return fmt.Errorf("failed to make matcher for plugin ''%s': %q", step.Plugin, err)
		}
		matchers[step.Plugin] = matcher
	}

	pes, err := getProjectEnvs(ctx, rc)
	if err != nil {
		return err
	}
	namespaces, err := rc.Environment().GetNamespaces(ctx, connect.NewRequest(&environment.GetNamespacesRequest{
		ProjectEnvs: pes,
	}))
	if err != nil {
		return err
	}

	results, err := getResults(ctx, rc, matchers, namespaces.Msg.GetNamespaces())
	if err != nil {
		return err
	}

	if base.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(&results, base.Flags.OutputType)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Project", "Environment", "Namespace", "Capsule", "Plugin")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, r := range results {
		tbl.AddRow(r.ProjectID, r.EnvironmentID, r.Namespace, r.CapsuleID, r.Plugin)
	}
	tbl.Print()

	return nil
}

func getOperatorConfig(
	ctx context.Context,
	operatorClient *base.OperatorClient,
	scheme *runtime.Scheme,
) (*v1alpha1.OperatorConfig, error) {
	var cfgYAML string
	if operatorConfig == "" {
		cfgResp, err := operatorClient.Capabilities.GetConfig(ctx, connect.NewRequest(&capabilities.GetConfigRequest{}))
		if err != nil {
			return nil, err
		}
		cfgYAML = cfgResp.Msg.GetYaml()
	} else {
		bytes, err := os.ReadFile(operatorConfig)
		if err != nil {
			return nil, err
		}
		cfgYAML = string(bytes)
	}

	// TODO Encapsulate config decoding
	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()
	config := &v1alpha1.OperatorConfig{}
	decodedConfig, _, err := decoder.Decode([]byte(cfgYAML), nil, config)
	if err != nil {
		return nil, err
	}
	var ok bool
	config, ok = decodedConfig.(*v1alpha1.OperatorConfig)
	if !ok {
		return nil, fmt.Errorf("decoded operator config had unexpected type %T", decodedConfig)
	}

	return config, nil
}

func getProjectEnvs(ctx context.Context, rc rig.Client) ([]*environment.ProjectEnvironment, error) {
	if len(projects) == 0 {
		resp, err := rc.Project().List(ctx, connect.NewRequest(&project.ListRequest{}))
		if err != nil {
			return nil, err
		}
		for _, p := range resp.Msg.GetProjects() {
			projects = append(projects, p.GetProjectId())
		}
	}

	if len(environments) == 0 {
		resp, err := rc.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
		if err != nil {
			return nil, err
		}
		for _, e := range resp.Msg.GetEnvironments() {
			environments = append(environments, e.GetEnvironmentId())
		}
	}
	var pes []*environment.ProjectEnvironment
	for _, p := range projects {
		for _, e := range environments {
			pes = append(pes, &environment.ProjectEnvironment{
				ProjectId:     p,
				EnvironmentId: e,
			})
		}
	}

	return pes, nil
}

type result struct {
	ProjectID     string `json:"project_id"`
	EnvironmentID string `json:"environment_id"`
	Namespace     string `json:"namespace"`
	CapsuleID     string `json:"capsule_id"`
	Plugin        string `json:"plugin"`
}

func getResults(
	ctx context.Context,
	rc rig.Client,
	matchers map[string]plugin.Matcher,
	namespaces []*environment.ProjectEnvironmentNamespace,
) ([]result, error) {
	projectMap := map[string][]*environment.ProjectEnvironmentNamespace{}
	for _, ns := range namespaces {
		projectMap[ns.GetProjectId()] = append(projectMap[ns.GetProjectId()], ns)
	}

	var results []result
	for pID, namespaces := range projectMap {
		var cs []string
		if len(capsules) > 0 {
			cs = capsules
		} else {
			resp, err := rc.Capsule().List(ctx, connect.NewRequest(&capsule.ListRequest{
				ProjectId: pID,
			}))
			if err != nil {
				return nil, err
			}
			for _, c := range resp.Msg.GetCapsules() {
				cs = append(cs, c.GetCapsuleId())
			}

		}

		for _, capsuleID := range cs {
			for _, ns := range namespaces {
				for plugin, matcher := range matchers {
					if matcher.Match(ns.GetNamespace(), capsuleID, nil) {
						results = append(results, result{
							ProjectID:     pID,
							EnvironmentID: ns.GetEnvironmentId(),
							Namespace:     ns.GetNamespace(),
							CapsuleID:     capsuleID,
							Plugin:        plugin,
						})
					}
				}
			}
		}
	}
	slices.SortFunc(results, func(r1, r2 result) int {
		if r1.ProjectID != r2.ProjectID {
			return strings.Compare(r1.ProjectID, r2.ProjectID)
		}
		if r1.EnvironmentID != r2.EnvironmentID {
			return strings.Compare(r1.EnvironmentID, r2.EnvironmentID)
		}
		if r1.Namespace != r2.Namespace {
			return strings.Compare(r1.Namespace, r2.Namespace)
		}
		if r1.CapsuleID != r2.CapsuleID {
			return strings.Compare(r1.CapsuleID, r2.CapsuleID)
		}
		return strings.Compare(r1.Plugin, r2.Plugin)
	})

	return results, nil
}
