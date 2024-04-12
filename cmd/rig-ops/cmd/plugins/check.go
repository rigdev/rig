package plugins

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) check(ctx context.Context, _ *cobra.Command, _ []string) error {
	cfg, err := base.GetOperatorConfig(ctx, c.OperatorClient, c.Scheme)
	if err != nil {
		return err
	}

	var matchers []plugin.Matcher
	for _, step := range cfg.Pipeline.Steps {
		matcher, err := plugin.NewMatcher(plugin.MatchFromStep(step))
		if err != nil {
			return fmt.Errorf("failed to make matcher for step '%v': %q", step, err)
		}
		matchers = append(matchers, matcher)
	}

	var objects []capsuleNamespace

	if len(capsules) == 0 || len(namespaces) == 0 {
		capsuleList := v1alpha2.CapsuleList{}
		if err := c.K8s.List(ctx, &capsuleList); err != nil {
			return err
		}
		for _, c := range capsuleList.Items {
			if len(capsules) != 0 && !slices.Contains(capsules, c.Name) {
				continue
			}
			if len(namespaces) != 0 && !slices.Contains(namespaces, c.Namespace) {
				continue
			}
			objects = append(objects, capsuleNamespace{
				namespace:   c.Namespace,
				capsule:     c.Name,
				annotations: c.Annotations,
			})
		}
	} else {
		for _, c := range capsules {
			for _, ns := range namespaces {
				objects = append(objects, capsuleNamespace{
					namespace: ns,
					capsule:   c,
				})
			}
		}
	}

	results, err := getResults(matchers, objects)
	if err != nil {
		return err
	}

	if base.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(&results, base.Flags.OutputType)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	tbl := table.
		New("Namespace", "Capsule", "StepIndex").
		WithHeaderFormatter(headerFmt)
	for _, r := range results {
		tbl.AddRow(r.Namespace, r.CapsuleID, r.StepIndex)
	}
	tbl.Print()

	return nil
}

type capsuleNamespace struct {
	namespace   string
	capsule     string
	annotations map[string]string
}

type result struct {
	Namespace string `json:"namespace"`
	CapsuleID string `json:"capsule_id"`
	StepIndex int    `json:"step_index"`
}

func getResults(matchers []plugin.Matcher, objects []capsuleNamespace) ([]result, error) {
	var results []result
	for _, obj := range objects {
		for idx, matcher := range matchers {
			if matcher.Match(obj.namespace, obj.capsule, obj.annotations) {
				results = append(results, result{
					Namespace: obj.namespace,
					CapsuleID: obj.capsule,
					StepIndex: idx,
				})
			}
		}
	}

	slices.SortFunc(results, func(r1, r2 result) int {
		if r1.Namespace != r2.Namespace {
			return strings.Compare(r1.Namespace, r2.Namespace)
		}
		if r1.CapsuleID != r2.CapsuleID {
			return strings.Compare(r1.CapsuleID, r2.CapsuleID)
		}
		return r1.StepIndex - r2.StepIndex
	})

	return results, nil
}
