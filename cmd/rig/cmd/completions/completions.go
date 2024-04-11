package completions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
)

func formatCapsule(c *capsule.Capsule) string {
	age := time.Since(c.GetUpdatedAt().AsTime()).Truncate(time.Second).String()

	return fmt.Sprintf("%v\t (Updated At: %v)", c.GetCapsuleId(), age)
}

func formatProject(p *project.Project) string {
	age := "-"
	if p.GetCreatedAt().IsValid() {
		age = p.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")
	}

	return fmt.Sprintf("%v\t (ID: %v, Created At: %v)", p.GetName(), p.GetProjectId(), age)
}

func formatEnvironment(e *environment.Environment) string {
	operatorVersion := "-"
	if e.GetOperatorVersion() != "" {
		operatorVersion = e.GetOperatorVersion()
	}

	return fmt.Sprintf("%v\t (Operator Version: %v)", e.GetEnvironmentId(), operatorVersion)
}

func Contexts(
	toComplete string,
	config *cmdconfig.Config,
) ([]string, cobra.ShellCompDirective) {
	names := []string{}

	for _, ctx := range config.Contexts {
		if strings.HasPrefix(ctx.Name, toComplete) {
			names = append(names, ctx.Name)
		}
	}

	if len(names) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func Environments(
	ctx context.Context,
	rig rig.Client,
	toComplete string,
	scope scope.Scope,
) ([]string, cobra.ShellCompDirective) {
	var environmentIDs []string

	if scope.GetCurrentContext() == nil || scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := rig.Environment().List(ctx, &connect.Request[environment.ListRequest]{
		Msg: &environment.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, p := range resp.Msg.GetEnvironments() {
		if strings.HasPrefix(p.GetEnvironmentId(), toComplete) {
			environmentIDs = append(environmentIDs, formatEnvironment(p))
		}
	}

	if len(environmentIDs) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return environmentIDs, cobra.ShellCompDirectiveNoFileComp
}

func Projects(
	ctx context.Context,
	rig rig.Client,
	toComplete string,
	scope scope.Scope,
) ([]string, cobra.ShellCompDirective) {
	var projectIDs []string

	if scope.GetCurrentContext() == nil || scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := rig.Project().List(ctx, &connect.Request[project.ListRequest]{
		Msg: &project.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, p := range resp.Msg.GetProjects() {
		if strings.HasPrefix(p.GetProjectId(), toComplete) {
			projectIDs = append(projectIDs, formatProject(p))
		}
	}

	if len(projectIDs) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return projectIDs, cobra.ShellCompDirectiveNoFileComp
}

func Capsules(
	ctx context.Context,
	rig rig.Client,
	toComplete string,
	scope scope.Scope,
) ([]string, cobra.ShellCompDirective) {
	var capsuleIDs []string

	if scope.GetCurrentContext() == nil || scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := rig.Capsule().List(ctx, &connect.Request[capsule.ListRequest]{
		Msg: &capsule.ListRequest{
			ProjectId: flags.GetProject(scope),
		},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, c := range resp.Msg.GetCapsules() {
		if strings.HasPrefix(c.GetCapsuleId(), toComplete) {
			capsuleIDs = append(capsuleIDs, formatCapsule(c))
		}
	}

	return capsuleIDs, cobra.ShellCompDirectiveDefault
}

func OutputType(_ *cobra.Command,
	_ []string,
	toComplete string) ([]string, cobra.ShellCompDirective) {
	options := []string{
		"json",
		"yaml",
		"pretty",
	}

	var completions []string

	for _, o := range options {
		if strings.HasPrefix(o, toComplete) {
			completions = append(completions, o)
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
