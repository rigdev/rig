package service_account

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	outputJSON bool
)

var (
	offset int
	limit  int
)

var (
	name string
)

type Cmd struct {
	fx.In

	Rig rig.Client
}

func Setup(parent *cobra.Command) {
	serviceAccount := &cobra.Command{
		Use:   "service-account",
		Short: "Manage service accounts",
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a new service account",
		RunE:  base.Register(func(c Cmd) any { return c.create }),
		Args:  cobra.NoArgs,
	}
	serviceAccount.PersistentFlags().StringVarP(&name, "name", "n", "", "name of the credential")
	serviceAccount.AddCommand(create)

	get := &cobra.Command{
		Use:               "get [id]",
		Short:             "Get one or multiple service accounts",
		RunE:              base.Register(func(c Cmd) any { return c.list }),
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: base.RegisterCompletion(func(c Cmd) any { return c.completions }),
	}
	get.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	get.Flags().IntVarP(&offset, "offset", "o", 0, "offset")
	get.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	get.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	get.RegisterFlagCompletionFunc("offset", common.NoCompletions)
	get.RegisterFlagCompletionFunc("limit", common.NoCompletions)

	serviceAccount.AddCommand(get)

	delete := &cobra.Command{
		Use:               "delete [id]",
		Short:             "Delete a service account",
		RunE:              base.Register(func(c Cmd) any { return c.delete }),
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: base.RegisterCompletion(func(c Cmd) any { return c.completions }),
	}
	serviceAccount.AddCommand(delete)

	parent.AddCommand(serviceAccount)
}

func (c Cmd) completions(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string
	accs, err := c.Rig.ServiceAccount().List(ctx, &connect.Request[service_account.ListRequest]{
		Msg: &service_account.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, acc := range accs.Msg.GetServiceAccounts() {
		if strings.HasPrefix(acc.GetServiceAccountId(), toComplete) {
			completions = append(completions, formatServiceAccount(acc))
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveDefault
}

func formatServiceAccount(acc *service_account.Entry) string {
	return fmt.Sprintf("%s\t (Name: %s)", acc.GetServiceAccountId(), acc.GetServiceAccount().GetName())
}
