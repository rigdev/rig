package user

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
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

	platform      string
	credFilePath  string
	usersFilePath string
	hashingKey    string

	groupIdentifier string
)

var (
	outputJson bool
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
}

func (c Cmd) Setup(parent *cobra.Command) {
	user := &cobra.Command{
		Use:   "user",
		Short: "Manage users in your projects",
	}

	create := &cobra.Command{
		Use:  "create",
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	create.Flags().StringVarP(&email, "email", "e", "", "email of the user")
	create.Flags().StringVarP(&username, "username", "u", "", "username of the user")
	create.Flags().StringVarP(&phoneNumber, "phone", "P", "", "phone number of the user")
	create.Flags().StringVarP(&password, "password", "p", "", "password of the user")
	user.AddCommand(create)

	update := &cobra.Command{
		Use:  "update [user-id | {email|username|phone}]",
		RunE: c.update,
		Args: cobra.MaximumNArgs(1),
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

	user.AddCommand(update)

	get := &cobra.Command{
		Use:  "get [user-id | {email|username|phone}]",
		RunE: c.lookup,
		Args: cobra.MaximumNArgs(1),
	}
	get.Flags().BoolVar(&outputJson, "json", false, "output as json")
	user.AddCommand(get)

	list := &cobra.Command{
		Use:  "list [search...]",
		RunE: c.list,
	}
	list.Flags().IntVarP(&offset, "offset", "o", 0, "offset for pagination")
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	list.Flags().BoolVar(&outputJson, "json", false, "output as json")
	user.AddCommand(list)

	delete := &cobra.Command{
		Use:  "delete [user-id | {email|username|phone}]",
		RunE: c.delete,
		Args: cobra.MaximumNArgs(1),
	}
	user.AddCommand(delete)

	listSessions := &cobra.Command{
		Use:  "list-sessions [user-id | {email|username|phone}]",
		RunE: c.listSessions,
		Args: cobra.MaximumNArgs(1),
	}
	listSessions.Flags().IntVarP(&offset, "offset", "o", 0, "offset for pagination")
	listSessions.Flags().IntVarP(&limit, "limit", "l", 10, "limit for pagination")
	listSessions.Flags().BoolVar(&outputJson, "json", false, "output as json")
	user.AddCommand(listSessions)

	getSettings := &cobra.Command{
		Use:  "get-settings",
		RunE: c.getSettings,
		Args: cobra.NoArgs,
	}
	getSettings.Flags().BoolVar(&outputJson, "json", false, "output as json")
	user.AddCommand(getSettings)

	updateSettings := &cobra.Command{
		Use:  "update-settings",
		RunE: c.updateSettings,
		Args: cobra.NoArgs,
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
					"  oauth-settings 		- json \n" +
					"  callbacks 			- json \n\n" +

					"Multi-Valued fields should be input as JSON \n"),
			)
		},
	)
	user.AddCommand(updateSettings)

	migrate := &cobra.Command{
		Use:  "migrate",
		RunE: c.migrate,
		Args: cobra.NoArgs,
	}
	migrate.Flags().StringVarP(&platform, "platform", "p", "Firebase", "platform to migrate from")
	migrate.Flags().StringVarP(&credFilePath, "cred-file", "c", "", "path to the credentials file")
	migrate.Flags().StringVarP(&usersFilePath, "users-file", "u", "", "path to the users file")
	migrate.Flags().StringVarP(&hashingKey, "hashing-key", "k", "", "key to use for hashing")
	migrate.MarkFlagsMutuallyExclusive("cred-file", "users-file")
	migrate.SetHelpFunc(
		func(cmd *cobra.Command, args []string) {
			cmd.Printf(
				("Usage:\n" +
					"  rig user migrate [flags] \n\n" +
					"Flags: \n" +
					"  -p, --platform string   platform to migrate from \n" +
					"  -c, --cred-file string  path to the credentials file \n" +
					"  -u, --users-file string path to the users file \n" +
					"  -h, --help 		 	  help for migrate \n" +
					"  -k, --hashing-key string key to use for hashing \n\n" +

					"Available platforms: \n" +
					"  Firebase \n\n" +

					"File paths should be absolute \n"),
			)
		},
	)
	user.AddCommand(migrate)

	addUser := &cobra.Command{
		Use:  "add-member [user-id | {email|username|phone}]",
		RunE: c.addMember,
		Args: cobra.MaximumNArgs(1),
	}
	addUser.Flags().StringVarP(&groupIdentifier, "group", "g", "", "group to add the user to")
	user.AddCommand(addUser)

	removeUser := &cobra.Command{
		Use:  "remove-member [user-id | {email|username|phone}]",
		RunE: c.removeMember,
		Args: cobra.MaximumNArgs(1),
	}
	removeUser.Flags().StringVarP(&groupIdentifier, "group", "g", "", "group to remove the user from")
	user.AddCommand(removeUser)

	parent.AddCommand(user)
}
