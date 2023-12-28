package serviceaccount

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var name string

type Cmd struct {
	fx.In

	Rig rig.Client
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
}

func Setup(parent *cobra.Command) {
	serviceAccount := &cobra.Command{
		Use:               "service-account",
		Short:             "Manage service accounts",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a new service account",
		RunE:  base.CtxWrap(cmd.create),
		Args:  cobra.NoArgs,
	}
	serviceAccount.PersistentFlags().StringVarP(&name, "name", "n", "", "name of the credential")
	serviceAccount.AddCommand(create)

	get := &cobra.Command{
		Use:               "get [id]",
		Short:             "Get one or multiple service accounts",
		RunE:              base.CtxWrap(cmd.list),
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: base.CtxWrapCompletion(cmd.completions),
	}
	get.Flags().IntVar(&offset, "offset", 0, "offset")
	get.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	serviceAccount.AddCommand(get)

	deleteCmd := &cobra.Command{
		Use:               "delete [id]",
		Short:             "Delete a service account",
		RunE:              base.CtxWrap(cmd.delete),
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: base.CtxWrapCompletion(cmd.completions),
	}
	serviceAccount.AddCommand(deleteCmd)

	parent.AddCommand(serviceAccount)
}

func (c *Cmd) completions(
	ctx context.Context,
	_ *cobra.Command,
	_ []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
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

func formatServiceAccount(acc *model.ServiceAccountEntry) string {
	return fmt.Sprintf("%s\t (Name: %s)", acc.GetServiceAccountId(), acc.GetName())
}
