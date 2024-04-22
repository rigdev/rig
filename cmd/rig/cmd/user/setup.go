package user

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
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

var (
	email    string
	password string
	role     string

	field string
	value string
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Prompter common.Prompter
	Scope    scope.Scope
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	user := &cobra.Command{
		Use:               "user",
		Short:             "Manage users",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
		},
		GroupID: common.AuthGroupID,
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		RunE:  cli.CtxWrap(cmd.create),
		Args:  cobra.NoArgs,
	}
	create.Flags().StringVarP(&email, "email", "e", "", "email of the user")
	create.Flags().StringVarP(&password, "password", "p", "", "password of the user")
	create.Flags().StringVarP(&role, "role", "r", "", "role of the user (admin, owner, developer, viewer)")
	if err := create.RegisterFlagCompletionFunc("role", common.RoleCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	user.AddCommand(create)

	update := &cobra.Command{
		Use:   "update [user-id | email]",
		Short: "Update a user",
		RunE:  cli.CtxWrap(cmd.update),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.userCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	update.Flags().StringVarP(&field, "field", "f", "", "field to update")
	update.Flags().StringVarP(&value, "value", "v", "", "value to update the field with")
	update.MarkFlagsRequiredTogether("field", "value")
	update.SetHelpFunc(
		func(cmd *cobra.Command, args []string) {
			cmd.Printf(
				("Usage:\n" +
					"  rig user update [user-id | email] [flags] \n\n" +
					"Flags: \n" +
					"  -f, --field string   field to update \n" +
					"  -v, --value string   value to update the field with \n" +
					"  -h, --help 		 	help for update \n\n" +

					"Available fields: \n" +
					"  email 		- string\n" +
					"  username		- string\n" +
					"  phone-number			- string\n" +
					"  profile		- json\n" +
					"  email-verified	- bool\n" +
					"  phone-verified	- bool\n" +
					"  set-meta-data		- json\n" +
					"  delete-meta-data	- string (key)\n"),
			)
		},
	)
	if err := update.RegisterFlagCompletionFunc("field", updateUserCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	user.AddCommand(update)

	list := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE:  cli.CtxWrap(cmd.list),
		Args:  cobra.NoArgs,
	}
	list.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	user.AddCommand(list)

	get := &cobra.Command{
		Use:   "get [user-id | email]",
		Short: "Get a user",
		RunE:  cli.CtxWrap(cmd.get),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.userCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	user.AddCommand(get)

	deleteCmd := &cobra.Command{
		Use:   "delete [user-id | email]",
		Short: "Delete a user",
		RunE:  cli.CtxWrap(cmd.delete),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.userCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	user.AddCommand(deleteCmd)

	listSessions := &cobra.Command{
		Use:   "list-sessions [user-id | email]",
		Short: "List sessions of a user",
		RunE:  cli.CtxWrap(cmd.listSessions),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.userCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	listSessions.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listSessions.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	user.AddCommand(listSessions)

	getSettings := &cobra.Command{
		Use:   "get-settings",
		Short: "Get the user-settings for the current project",
		RunE:  cli.CtxWrap(cmd.getSettings),
		Args:  cobra.NoArgs,
	}
	user.AddCommand(getSettings)

	updateSettings := &cobra.Command{
		Use:   "update-settings",
		Short: "Update the user-settings for the current project",
		RunE:  cli.CtxWrap(cmd.updateSettings),
		Args:  cobra.NoArgs,
	}
	updateSettings.Flags().StringVarP(&field, "field", "f", "", "field to update")
	updateSettings.Flags().StringVarP(&value, "value", "v", "", "value to update the field with")
	updateSettings.MarkFlagsRequiredTogether("field", "value")
	updateSettings.SetHelpFunc(
		func(cmd *cobra.Command, args []string) {
			cmd.Printf(
				("Usage:\n" +
					"  rig user update-settings [flags] \n\n" +
					"Flags: \n" +
					"  -f, --field string   field to update \n" +
					"  -v, --value string   value to update the field with \n" +
					"  -h, --help 		 	help for update-settings \n\n" +

					"Available fields: \n" +
					"  allow-register 		- bool\n" +
					"  allow-login 			- bool\n" +
					"  verify-email-required 	- bool\n" +
					"  verify-phone-required 	- bool \n" +
					"  access-token-ttl 		- int (minutes) \n" +
					"  refresh-token-ttl 		- int (hours) \n" +
					"  verification-code-ttl 	- int (minutes) \n" +
					"  password-hashing 		- json \n" +
					"  login-mechanisms 		- json \n" +
					"  email-provider 		- json \n" +
					"  template 			- json \n\n" +

					"Multi-Valued fields should be input as JSON \n"),
			)
		},
	)
	if err := updateSettings.RegisterFlagCompletionFunc("field", updateSettingsCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	user.AddCommand(updateSettings)

	parent.AddCommand(user)
}

func (c *Cmd) userCompletions(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.User().List(ctx, &connect.Request[user.ListRequest]{
		Msg: &user.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	for _, u := range resp.Msg.GetUsers() {
		if strings.HasPrefix(u.GetPrintableName(), toComplete) {
			completions = append(completions, formatUser(u))
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func updateSettingsCompletions(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	options := []string{
		"allow-register",
		"allow-login",
		"verify-email-required",
		"verify-phone-required",
		"access-token-ttl",
		"refresh-token-ttl",
		"verification-code-ttl",
		"password-hashing",
		"login-mechanisms",
		"email-provider",
		"template",
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

	return completions, cobra.ShellCompDirectiveDefault
}

func updateUserCompletions(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fields := []string{
		"email",
		"username",
		"phone-number",
		"profile",
		"email-verified",
		"phone-verified",
		"set-meta-data",
		"delete-meta-data",
	}

	var completions []string

	for _, f := range fields {
		if strings.HasPrefix(f, toComplete) {
			completions = append(completions, f)
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveDefault
}

func formatUser(u *model.UserEntry) string {
	return fmt.Sprintf("%s\t (ID: %s)", u.GetPrintableName(), u.GetUserId())
}
