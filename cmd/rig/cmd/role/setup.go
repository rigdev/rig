package role

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/role"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Prompter common.Prompter
	Scope    scope.Scope
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Scope = c.Scope
	cmd.Prompter = c.Prompter
}

var (
	limit  int
	offset int
)

var (
	roleType    string
	project     string
	environment string

	updateRoleType    string
	updateProject     string
	updateEnvironment string
)

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	roles := &cobra.Command{
		Use:               "role",
		Short:             "Manage roles",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
		},
		GroupID: common.AuthGroupID,
	}

	create := &cobra.Command{
		Use:   "create [role-id]",
		Short: "Create a new role",
		RunE:  cli.CtxWrap(cmd.create),
		Args:  cobra.MaximumNArgs(1),
	}
	create.Flags().StringVarP(&roleType, "type", "t", "owner",
		"Select the role type to create. Must be one of (admin, owner, developer, viewer)")
	create.Flags().StringVar(&project, "project", "*", "Select the project to give the roll access to. "+
		"If none is provided, the role will have access to all projects")
	create.Flags().StringVar(&environment, "environment", "*", "Select the environment to give the role access to. "+
		"If none is provided, the role will have access to all environments")

	if err := create.RegisterFlagCompletionFunc("type", common.RoleCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := create.RegisterFlagCompletionFunc("project",
		cli.HackCtxWrapCompletion(cmd.completeProject, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := create.RegisterFlagCompletionFunc("environment",
		cli.HackCtxWrapCompletion(cmd.completeEnvironment, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	roles.AddCommand(create)

	deleteCmd := &cobra.Command{
		Use:   "delete [role-id]",
		Short: "Delete a role",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.completeRole, s),
			common.MaxArgsCompletionFilter(1),
		),
		RunE: cli.CtxWrap(cmd.delete),
	}
	roles.AddCommand(deleteCmd)

	list := &cobra.Command{
		Use:   "list",
		Short: "List roles",
		RunE:  cli.CtxWrap(cmd.list),
		Args:  cobra.NoArgs,
	}
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	list.Flags().IntVar(&offset, "offset", 0, "offset the number of groups to return")
	roles.AddCommand(list)

	listRolesForGroup := &cobra.Command{
		Use:   "list-for-group [group-id]",
		Short: "List roles for a group",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.listRolesForGroup),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.completeGroup, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	listRolesForGroup.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	listRolesForGroup.Flags().IntVar(&offset, "offset", 0, "offset the number of groups to return")
	roles.AddCommand(listRolesForGroup)

	listGroupsForRole := &cobra.Command{
		Use:   "list-for-role [role-id]",
		Short: "List groups for a role",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.listGroupsForRole),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.completeRole, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	listGroupsForRole.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	listGroupsForRole.Flags().IntVar(&offset, "offset", 0, "offset the number of groups to return")
	roles.AddCommand(listGroupsForRole)

	assign := &cobra.Command{
		Use:   "assign [role-id] [group-id]",
		Short: "Assign a role to a group",
		RunE:  cli.CtxWrap(cmd.assign),
		Args:  cobra.MaximumNArgs(2),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.completeRole, s),
			cli.HackCtxWrapCompletion(cmd.completeGroup, s),
		),
	}
	roles.AddCommand(assign)

	revoke := &cobra.Command{
		Use:   "revoke [role-id] [group-id]",
		Short: "Revoke a role from a group",
		RunE:  cli.CtxWrap(cmd.revoke),
		Args:  cobra.MaximumNArgs(2),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.completeRole, s),
			cli.HackCtxWrapCompletion(cmd.completeGroup, s),
		),
	}
	roles.AddCommand(revoke)

	get := &cobra.Command{
		Use:   "get [role-id]",
		Short: "Get a role",
		RunE:  cli.CtxWrap(cmd.get),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.completeRole, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	roles.AddCommand(get)

	update := &cobra.Command{
		Use:   "update [role-id]",
		Short: "Update a role",
		RunE:  cli.CtxWrap(cmd.update),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.completeRole, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	update.Flags().StringVarP(&updateRoleType, "type", "t", "",
		"Select the role type to update. Must be one of (admin, owner, developer, viewer)")
	update.Flags().StringVar(&updateProject, "project", "", "Select the project to give the roll access to. "+
		"If none is provided, the role will have access to all projects")
	update.Flags().StringVar(&updateEnvironment, "environment", "", "Select the environment to give the role access to. "+
		"If none is provided, the role will have access to all environments")

	if err := update.RegisterFlagCompletionFunc("type", common.RoleCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := update.RegisterFlagCompletionFunc("project",
		cli.HackCtxWrapCompletion(cmd.completeProject, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := update.RegisterFlagCompletionFunc("environment",
		cli.HackCtxWrapCompletion(cmd.completeEnvironment, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	roles.AddCommand(update)

	parent.AddCommand(roles)
}

func formatRole(r *role.Role) string {
	roleType := r.GetRoleId()
	if t, ok := r.GetMetadata()["roleType"]; ok {
		roleType = string(t)
	}

	project := r.GetPermissions()[0].GetScope().GetProject()
	environment := r.GetPermissions()[0].GetScope().GetEnvironment()

	return fmt.Sprintf("%s\t (Type: %s, Project: %s, Environment: %s)", r.GetRoleId(), roleType, project, environment)
}

func (c *Cmd) completeRole(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	rolesResp, err := c.Rig.Role().List(ctx, connect.NewRequest(
		&role.ListRequest{},
	))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	roles := []string{}
	for _, r := range rolesResp.Msg.GetRoles() {
		if strings.HasPrefix(r.GetRoleId(), toComplete) {
			roles = append(roles, formatRole(r))
		}
	}

	return roles, cobra.ShellCompDirectiveNoFileComp
}

func (c *Cmd) completeGroup(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions.Groups(ctx, c.Rig, toComplete)
}

func (c *Cmd) completeProject(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions.Projects(ctx, c.Rig, toComplete)
}

func (c *Cmd) completeEnvironment(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions.Environments(ctx, c.Rig, toComplete)
}
