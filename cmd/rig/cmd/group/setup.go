package group

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
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

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	group := &cobra.Command{
		Use:   "group",
		Short: "Manage role groups",
		Long: "Groups are a way to organize users and service accounts into groups with certain roles, where " +
			"the roles assigned to a group are inherited by all members of the group.",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
		},
		GroupID: common.AuthGroupID,
	}

	create := &cobra.Command{
		Use:   "create [group-id]",
		Short: "Create a new group",
		RunE:  cli.CtxWrap(cmd.create),
		Args:  cobra.MaximumNArgs(1),
	}
	group.AddCommand(create)

	deleteCmd := &cobra.Command{
		Use:   "delete [group-id]",
		Short: "Delete a group",
		RunE:  cli.CtxWrap(cmd.delete),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	group.AddCommand(deleteCmd)

	update := &cobra.Command{
		Use:   "update [group-id]",
		Short: "Update a group",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.update),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	group.AddCommand(update)

	list := &cobra.Command{
		Use:     "list",
		Short:   "list groups",
		Aliases: []string{"ls"},
		RunE:    cli.CtxWrap(cmd.list),
	}
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	list.Flags().IntVar(&offset, "offset", 0, "offset the number of groups to return")
	group.AddCommand(list)

	listMembers := &cobra.Command{
		Use:   "list-members [group-id]",
		Short: "list members of a group",
		RunE:  cli.CtxWrap(cmd.listMembers),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.completions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	listMembers.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of members to return")
	listMembers.Flags().IntVar(&offset, "offset", 0, "offset the number of members to return")
	group.AddCommand(listMembers)

	listGroupsForMember := &cobra.Command{
		Use:   "list-groups-for-member [member-id]",
		Short: "List groups that a user or service account is a member of",
		RunE:  cli.CtxWrap(cmd.listGroupsForUser),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.memberCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	listGroupsForMember.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	listGroupsForMember.Flags().IntVar(&offset, "offset", 0, "offset the number of groups to return")
	group.AddCommand(listGroupsForMember)

	addUser := &cobra.Command{
		Use:   "add-member [member-id] [group-id]",
		Short: "Add a member to a group",
		RunE:  cli.CtxWrap(cmd.addMember),
		Args:  cobra.MaximumNArgs(2),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.memberCompletions, s),
			cli.HackCtxWrapCompletion(cmd.completions, s),
		),
	}
	group.AddCommand(addUser)

	removeUser := &cobra.Command{
		Use:   "remove-member [member-id] [group-id]",
		Short: "Remove a member from a group",
		RunE:  cli.CtxWrap(cmd.removeMember),
		Args:  cobra.MaximumNArgs(2),
		ValidArgsFunction: common.ChainCompletions(
			[]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.memberCompletions, s),
			cli.HackCtxWrapCompletion(cmd.completions, s),
		),
	}
	group.AddCommand(removeUser)

	parent.AddCommand(group)
}

func (c *Cmd) completions(
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

func (c *Cmd) memberCompletions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	completions := []string{}
	resp, err := c.Rig.User().List(ctx, &connect.Request[user.ListRequest]{
		Msg: &user.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	saResp, err := c.Rig.ServiceAccount().List(ctx, &connect.Request[service_account.ListRequest]{
		Msg: &service_account.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, u := range resp.Msg.GetUsers() {
		if strings.HasPrefix(u.GetUserId(), toComplete) {
			completions = append(completions, formatUser(u))
		}
	}

	for _, sa := range saResp.Msg.GetServiceAccounts() {
		if strings.HasPrefix(sa.GetServiceAccountId(), toComplete) {
			completions = append(completions, formatServiceAccount(sa))
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func formatUser(u *model.UserEntry) string {
	return fmt.Sprintf("%s\t (User: %s)", u.GetUserId(), u.GetPrintableName())
}

func formatServiceAccount(u *model.ServiceAccountEntry) string {
	return fmt.Sprintf("%s\t (SA: %s)", u.GetServiceAccountId(), u.GetName())
}
