package service_account

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
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

	Ctx context.Context
	Rig rig.Client
}

func (c Cmd) Setup(parent *cobra.Command) {
	serviceAccount := &cobra.Command{
		Use:   "service-account",
		Short: "Manage service accounts",
	}

	create := &cobra.Command{
		Use:  "create",
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	create.Flags().StringVarP(&name, "name", "n", "", "name of the credential")
	serviceAccount.AddCommand(create)

	list := &cobra.Command{
		Use:  "list",
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	list.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	list.Flags().IntVarP(&offset, "offset", "o", 0, "offset")
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit")

	serviceAccount.AddCommand(list)

	delete := &cobra.Command{
		Use:  "delete [id]",
		RunE: c.delete,
		Args: cobra.MaximumNArgs(1),
	}
	serviceAccount.AddCommand(delete)

	parent.AddCommand(serviceAccount)
}
