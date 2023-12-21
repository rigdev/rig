package group

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	groupID string
)

type Cmd struct {
	fx.In

	Rig rig.Client
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
}

func Setup(parent *cobra.Command) {
	group := &cobra.Command{
		Use:               "group",
		Short:             "Manage groups",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a new group",
		RunE:  base.CtxWrap(cmd.create),
		Args:  cobra.NoArgs,
	}
	create.Flags().StringVarP(&groupID, "group-id", "g", "", "id of the group")
	group.AddCommand(create)

	deleteCmd := &cobra.Command{
		Use:   "delete [group-id]",
		Short: "Delete a group",
		RunE:  base.CtxWrap(cmd.delete),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	group.AddCommand(deleteCmd)

	update := &cobra.Command{
		Use:   "update [group-id]",
		Short: "Update a group",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.update),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	group.AddCommand(update)

	get := &cobra.Command{
		Use:     "get [group-id]",
		Short:   "Get one or multiple groups",
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"ls"},
		RunE:    base.CtxWrap(cmd.get),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	get.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	get.Flags().IntVar(&offset, "offset", 0, "offset the number of groups to return")
	group.AddCommand(get)

	getMembers := &cobra.Command{
		Use:   "get-members [group-id]",
		Short: "Get members of a group",
		RunE:  base.CtxWrap(cmd.listMembers),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.completions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	getMembers.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of members to return")
	getMembers.Flags().IntVar(&offset, "offset", 0, "offset the number of members to return")
	group.AddCommand(getMembers)

	getGroupsForMember := &cobra.Command{
		Use:   "get-groups-for-member [member-id]",
		Short: "Get groups that a user or service account is a member of",
		RunE:  base.CtxWrap(cmd.listGroupsForUser),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.memberCompletions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	getGroupsForMember.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	getGroupsForMember.Flags().IntVar(&offset, "offset", 0, "offset the number of groups to return")
	group.AddCommand(getGroupsForMember)

	addUser := &cobra.Command{
		Use:   "add-member [member-id]",
		Short: "Add a member to a group",
		RunE:  base.CtxWrap(cmd.addMember),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.memberCompletions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	addUser.Flags().StringVarP(&groupID, "group-id", "g", "", "id of the group")
	if err := addUser.RegisterFlagCompletionFunc(
		"group-id",
		base.CtxWrapCompletion(cmd.completions),
	); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	group.AddCommand(addUser)

	removeUser := &cobra.Command{
		Use:   "remove-member [member-id]",
		Short: "Remove a member from a group",
		RunE:  base.CtxWrap(cmd.removeMember),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.memberCompletions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	removeUser.Flags().StringVarP(&groupID, "group-id", "g", "", "id of the group")
	if err := removeUser.RegisterFlagCompletionFunc(
		"group-id",
		base.CtxWrapCompletion(cmd.completions),
	); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	group.AddCommand(removeUser)

	parent.AddCommand(group)
}

func (c *Cmd) completions(
	ctx context.Context,
	_ *cobra.Command,
	_ []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	completions := []string{}
	resp, err := c.Rig.Group().List(ctx, &connect.Request[group.ListRequest]{
		Msg: &group.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, g := range resp.Msg.GetGroups() {
		if strings.HasPrefix(g.GetGroupId(), toComplete) {
			completions = append(completions, formatGroup(g))
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func formatGroup(g *group.Group) string {
	return fmt.Sprintf("%s\t (#Members: %v)", g.GetGroupId(), g.GetNumMembers())
}

func (c *Cmd) memberCompletions(
	ctx context.Context,
	_ *cobra.Command,
	_ []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
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
	return fmt.Sprintf("%s\t (%s)", u.GetUserId(), u.GetPrintableName())
}

func formatServiceAccount(u *model.ServiceAccountEntry) string {
	return fmt.Sprintf("%s\t (%s)", u.GetServiceAccountId(), u.GetName())
}
