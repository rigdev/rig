package serviceaccount

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	name string
	role string
)

type Cmd struct {
	fx.In

	Rig rig.Client
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	serviceAccount := &cobra.Command{
		Use:               "service-account",
		Short:             "Manage service accounts",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
		},
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a new service account",
		RunE:  cli.CtxWrap(cmd.create),
		Args:  cobra.NoArgs,
	}
	serviceAccount.PersistentFlags().StringVarP(&name, "name", "n", "", "name of the credential")
	create.Flags().StringVarP(&role, "role", "r", "", "role of the user (admin, owner, developer, viewer)")
	if err := create.RegisterFlagCompletionFunc("role", common.RoleCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	serviceAccount.AddCommand(create)

	get := &cobra.Command{
		Use:               "get [id]",
		Short:             "Get one or multiple service accounts",
		RunE:              cli.CtxWrap(cmd.list),
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: cli.HackCtxWrapCompletion(cmd.completions, s),
	}
	get.Flags().IntVar(&offset, "offset", 0, "offset")
	get.Flags().IntVarP(&limit, "limit", "l", 10, "limit")
	serviceAccount.AddCommand(get)

	deleteCmd := &cobra.Command{
		Use:               "delete [id]",
		Short:             "Delete a service account",
		RunE:              cli.CtxWrap(cmd.delete),
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: cli.HackCtxWrapCompletion(cmd.completions, s),
	}
	serviceAccount.AddCommand(deleteCmd)

	parent.AddCommand(serviceAccount)
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
