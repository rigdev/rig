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
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	email       string
	username    string
	phoneNumber string
	password    string

	field string
	value string
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
	user := &cobra.Command{
		Use:               "user",
		Short:             "Manage users in your projects",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
		},
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		RunE:  base.CtxWrap(cmd.create),
		Args:  cobra.NoArgs,
	}
	create.Flags().StringVarP(&email, "email", "E", "", "email of the user")
	create.Flags().StringVarP(&username, "username", "u", "", "username of the user")
	create.Flags().StringVarP(&phoneNumber, "phone", "P", "", "phone number of the user")
	create.Flags().StringVarP(&password, "password", "p", "", "password of the user")
	user.AddCommand(create)

	update := &cobra.Command{
		Use:   "update [user-id | {email|username|phone}]",
		Short: "Update a user",
		RunE:  base.CtxWrap(cmd.update),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.userCompletions),
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
					"  rig user update [user-id | {email|username|phone}] [flags] \n\n" +
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

	get := &cobra.Command{
		Use:   "get [user-id | {email|username|phone}]",
		Short: "Get one or multiple users",
		RunE:  base.CtxWrap(cmd.get),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.userCompletions),
			common.MaxArgsCompletionFilter(1)),
		Args: cobra.MaximumNArgs(1),
	}
	get.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	get.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	user.AddCommand(get)

	deleteCmd := &cobra.Command{
		Use:   "delete [user-id | {email|username|phone}]",
		Short: "Delete a user",
		RunE:  base.CtxWrap(cmd.delete),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.userCompletions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	user.AddCommand(deleteCmd)

	getSessions := &cobra.Command{
		Use:   "get-sessions [user-id | {email|username|phone}]",
		Short: "Get sessions of a user",
		RunE:  base.CtxWrap(cmd.listSessions),
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			base.CtxWrapCompletion(cmd.userCompletions),
			common.MaxArgsCompletionFilter(1),
		),
	}
	getSessions.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	getSessions.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	user.AddCommand(getSessions)

	getSettings := &cobra.Command{
		Use:   "get-settings",
		Short: "Get the user-settings for the current project",
		RunE:  base.CtxWrap(cmd.getSettings),
		Args:  cobra.NoArgs,
	}
	user.AddCommand(getSettings)

	updateSettings := &cobra.Command{
		Use:   "update-settings",
		Short: "Update the user-settings for the current project",
		RunE:  base.CtxWrap(cmd.updateSettings),
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
) ([]string, cobra.ShellCompDirective) {
	if err := base.Provide(cmd, args, initCmd); err != nil {
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
