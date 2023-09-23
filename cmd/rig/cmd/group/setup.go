package group

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
	group := &cobra.Command{
		Use:   "group",
		Short: "Manage user groups",
	}

	create := &cobra.Command{
		Use:  "create",
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	create.Flags().StringVarP(&name, "name", "n", "", "name of the group")
	group.AddCommand(create)

	delete := &cobra.Command{
		Use:  "delete [group-id | group-name]",
		RunE: c.delete,
		Args: cobra.MaximumNArgs(1),
	}
	group.AddCommand(delete)

	update := &cobra.Command{
		Use:  "update [group-id | group-name]",
		Args: cobra.MaximumNArgs(1),
		RunE: c.update,
	}
	group.AddCommand(update)

	get := &cobra.Command{
		Use:  "get [group-id | group-name]",
		Args: cobra.MaximumNArgs(1),
		RunE: c.get,
	}
	get.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	group.AddCommand(get)

	list := &cobra.Command{
		Use:     "list [search...]",
		Args:    cobra.MinimumNArgs(0),
		Aliases: []string{"ls"},
		RunE:    c.list,
	}
	list.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	list.Flags().IntVarP(&offset, "offset", "o", 0, "offset the number of groups to return")
	group.AddCommand(list)

	listMembers := &cobra.Command{
		Use:  "list-members [group-id | group-name]",
		RunE: c.listMembers,
		Args: cobra.MaximumNArgs(1),
	}
	listMembers.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	listMembers.PersistentFlags().IntVarP(&limit, "limit", "l", 10, "limit the number of members to return")
	listMembers.PersistentFlags().IntVarP(&offset, "offset", "o", 0, "offset the number of members to return")
	group.AddCommand(listMembers)

	listGroupsForUser := &cobra.Command{
		Use:  "list-groups-for-user [user-id | {email|username|phone}]",
		RunE: c.listGroupsForUser,
		Args: cobra.MaximumNArgs(1),
	}
	listGroupsForUser.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	listGroupsForUser.PersistentFlags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	listGroupsForUser.PersistentFlags().IntVarP(&offset, "offset", "o", 0, "offset the number of groups to return")
	group.AddCommand(listGroupsForUser)

	parent.AddCommand(group)
}
