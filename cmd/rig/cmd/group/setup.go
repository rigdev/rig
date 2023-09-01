package group

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
	group := &cobra.Command{
		Use: "group",
	}

	create := &cobra.Command{
		Use:  "create",
		RunE: base.Register(GroupCreate),
		Args: cobra.NoArgs,
	}
	create.Flags().StringVarP(&name, "name", "n", "", "name of the group")
	group.AddCommand(create)

	delete := &cobra.Command{
		Use:  "delete [group-id | group-name]",
		RunE: base.Register(GroupDelete),
		Args: cobra.MaximumNArgs(1),
	}
	group.AddCommand(delete)

	update := &cobra.Command{
		Use:  "update [group-id | group-name]",
		Args: cobra.MaximumNArgs(1),
		RunE: base.Register(GroupUpdate),
	}
	group.AddCommand(update)

	get := &cobra.Command{
		Use:  "get [group-id | group-name]",
		Args: cobra.MaximumNArgs(1),
		RunE: base.Register(GroupGet),
	}
	get.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	group.AddCommand(get)

	list := &cobra.Command{
		Use:     "list [search...]",
		Args:    cobra.MinimumNArgs(0),
		Aliases: []string{"ls"},
		RunE:    base.Register(GroupList),
	}
	list.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	list.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	list.Flags().IntVarP(&offset, "offset", "o", 0, "offset the number of groups to return")
	group.AddCommand(list)

	listMembers := &cobra.Command{
		Use:  "list-members [group-id | group-name]",
		RunE: base.Register(GroupListMembers),
		Args: cobra.MaximumNArgs(1),
	}
	listMembers.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	listMembers.PersistentFlags().IntVarP(&limit, "limit", "l", 10, "limit the number of members to return")
	listMembers.PersistentFlags().IntVarP(&offset, "offset", "o", 0, "offset the number of members to return")
	group.AddCommand(listMembers)

	listGroupsForUser := &cobra.Command{
		Use:  "list-groups-for-user [user-id | {email|username|phone}]",
		RunE: base.Register(GroupListGroupsForUser),
		Args: cobra.MaximumNArgs(1),
	}
	listGroupsForUser.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	listGroupsForUser.PersistentFlags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	listGroupsForUser.PersistentFlags().IntVarP(&offset, "offset", "o", 0, "offset the number of groups to return")
	group.AddCommand(listGroupsForUser)

	parent.AddCommand(group)
}
