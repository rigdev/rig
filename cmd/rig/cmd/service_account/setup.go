package service_account

import (
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
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

func Setup(parent *cobra.Command) {
	serviceAccount := &cobra.Command{
		Use:   "service-account",
		Short: "Manage service accounts",
	}

	create := &cobra.Command{
		Use:  "create",
		RunE: base.Register(ServiceAccountCreate),
		Args: cobra.NoArgs,
	}
	create.Flags().StringVarP(&name, "name", "n", "", "name of the credential")
	serviceAccount.AddCommand(create)

	list := &cobra.Command{
		Use:  "list",
		RunE: base.Register(ServiceAccountList),
		Args: cobra.NoArgs,
	}
	list.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	list.Flags().IntVarP(&offset, "offset", "o", 0, "offset")
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit")

	serviceAccount.AddCommand(list)

	delete := &cobra.Command{
		Use:  "delete [id]",
		RunE: base.Register(ServiceAccountDelete),
		Args: cobra.MaximumNArgs(1),
	}
	serviceAccount.AddCommand(delete)

	parent.AddCommand(serviceAccount)
}
