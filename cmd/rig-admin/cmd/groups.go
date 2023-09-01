package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/lucasepe/codename"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	service_auth "github.com/rigdev/rig/internal/service/auth"
	group_service "github.com/rigdev/rig/internal/service/group"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	groupName  string
	groupCount int

	groupOffset int
	groupLimit  int
)

func init() {
	groups := &cobra.Command{
		Use: "groups",
	}

	populate := &cobra.Command{
		Use:  "populate",
		RunE: register(GroupsPopulate),
	}
	populate.PersistentFlags().IntVarP(&groupCount, "count", "n", 1, "number of groups to create")
	populate.PersistentFlags().StringVar(&groupName, "name", "", "name for the group")
	groups.AddCommand(populate)

	update := &cobra.Command{
		Use:  "update <group-id>",
		Args: cobra.MinimumNArgs(1),
		RunE: register(GroupsUpdate),
	}
	update.PersistentFlags().StringVar(&groupName, "name", "", "name for the group")
	groups.AddCommand(update)

	delete := &cobra.Command{
		Use:  "delete <group-id>",
		RunE: register(GroupsDelete),
		Args: cobra.ExactArgs(1),
	}
	groups.AddCommand(delete)

	get := &cobra.Command{
		Use:  "get <group-id>",
		RunE: register(GroupsGet),
		Args: cobra.ExactArgs(1),
	}
	groups.AddCommand(get)

	list := &cobra.Command{
		Use:  "list",
		RunE: register(GroupsList),
	}
	list.PersistentFlags().IntVarP(&groupLimit, "limit", "l", 10, "limit the number of groups to return")
	list.PersistentFlags().IntVarP(&groupOffset, "offset", "o", 0, "offset the number of groups to return")
	groups.AddCommand(list)

	addUser := &cobra.Command{
		Use:  "add-user <group-id> <user-id>",
		RunE: register(GroupAddUser),
		Args: cobra.ExactArgs(2),
	}
	groups.AddCommand(addUser)

	removeUser := &cobra.Command{
		Use:  "remove-user <group-id> <user-id>",
		RunE: register(GroupRemoveUser),
		Args: cobra.ExactArgs(2),
	}
	groups.AddCommand(removeUser)

	listMembers := &cobra.Command{
		Use:  "list-members <group-id>",
		RunE: register(GroupListMembers),
		Args: cobra.MinimumNArgs(1),
	}
	listMembers.PersistentFlags().IntVarP(&groupLimit, "limit", "l", 10, "limit the number of groups to return")
	listMembers.PersistentFlags().IntVarP(&groupOffset, "offset", "o", 0, "offset the number of groups to return")
	groups.AddCommand(listMembers)

	rootCmd.AddCommand(groups)
}

func GroupsPopulate(ctx context.Context, cmd *cobra.Command, gp *group_service.Service, logger *zap.Logger) error {
	logger.Info("created groups", zap.Int("count", groupCount))

	ctx = auth.WithClaims(ctx, service_auth.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	for i := 0; i < groupCount; i++ {

		var ups []*group.Update

		rng, err := codename.DefaultRNG()
		if err != nil {
			return err
		}

		ups = append(ups, &group.Update{Field: &group.Update_Name{Name: strings.ToLower(codename.Generate(rng, 0))}})

		g, err := gp.Create(ctx, ups)
		if err != nil {
			return err
		}

		logger.Info("created group", zap.String("group_id", g.GetGroupId()), zap.String("name", g.GetName()))
	}

	return nil
}

func GroupsList(ctx context.Context, cmd *cobra.Command, gp *group_service.Service, logger *zap.Logger) error {
	ctx = auth.WithClaims(ctx, service_auth.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	it, total, err := gp.List(ctx, &model.Pagination{
		Offset: uint32(groupOffset),
		Limit:  uint32(groupLimit),
	}, "")
	if err != nil {
		return err
	}
	logger.Info("groups listed", zap.Int("total", int(total)))
	for {
		g, err := it.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		logger.Info("group listed", zap.String("name", g.GetName()), zap.String("group_id", g.GetGroupId()), zap.Int("num_members", int(g.GetNumMembers())))
	}
	return nil
}

func GroupListForUser(ctx context.Context, cmd *cobra.Command, args []string, gp *group_service.Service, logger *zap.Logger) error {
	ctx = auth.WithClaims(ctx, service_auth.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	userID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	it, total, err := gp.ListGroupsForUser(ctx, userID, &model.Pagination{
		Offset: uint32(groupOffset),
		Limit:  uint32(groupLimit),
	})
	if err != nil {
		return err
	}

	logger.Info("groups listed", zap.Int("total", int(total)))
	for {
		g, err := it.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		logger.Info("group listed", zap.String("name", g.GetName()), zap.String("group_id", g.GetGroupId()), zap.Int("num_members", int(g.GetNumMembers())))
	}
	return nil
}

func GroupAddUser(ctx context.Context, cmd *cobra.Command, args []string, gp *group_service.Service, logger *zap.Logger) error {
	groupID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(args[1])
	if err != nil {
		return err
	}

	ctx = auth.WithClaims(ctx, service_auth.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	err = gp.AddMembers(ctx, groupID, []uuid.UUID{userID})
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("added user %v to group %v", args[1], args[0]))
	return nil
}

func GroupsDelete(ctx context.Context, cmd *cobra.Command, args []string, gp *group_service.Service, logger *zap.Logger) error {
	ctx = auth.WithClaims(ctx, service_auth.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	groupID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	err = gp.Delete(ctx, groupID)
	if err != nil {
		return err
	}

	logger.Info("group deleted")
	return nil
}

func GroupListMembers(ctx context.Context, cmd *cobra.Command, args []string, gp *group_service.Service, logger *zap.Logger) error {
	ctx = auth.WithClaims(ctx, service_auth.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	groupID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	it, total, err := gp.ListMembers(ctx, groupID, &model.Pagination{
		Offset: uint32(groupOffset),
		Limit:  uint32(groupLimit),
	})
	if err != nil {
		return err
	}
	logger.Info("group members listed", zap.Int("total", int(total)))
	for {
		u, err := it.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		logger.Info("group member listed", zap.String("name", u.GetUser().GetPrintableName()), zap.String("user_id", u.GetUser().GetUserId()))
	}
	return nil
}

func GroupsGet(ctx context.Context, cmd *cobra.Command, args []string, gp *group_service.Service, logger *zap.Logger) error {
	ctx = auth.WithClaims(ctx, service_auth.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	groupID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	g, err := gp.Get(ctx, groupID)
	if err != nil {
		return err
	}

	logger.Info("group listed", zap.String("name", g.GetName()), zap.Int("num_members", int(g.GetNumMembers())))
	return nil
}

func GroupsUpdate(ctx context.Context, cmd *cobra.Command, args []string, gp *group_service.Service, logger *zap.Logger) error {
	ctx = auth.WithClaims(ctx, service_auth.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	groupID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	var ups []*group.Update

	if groupName != "" {
		ups = append(ups, &group.Update{Field: &group.Update_Name{Name: groupName}})
	}

	err = gp.Update(ctx, groupID, ups)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("group %s updated", args[0]), zap.String("name: ", groupName))
	return nil
}

func GroupRemoveUser(ctx context.Context, cmd *cobra.Command, args []string, gp *group_service.Service, logger *zap.Logger) error {
	groupID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(args[1])
	if err != nil {
		return err
	}
	ctx = auth.WithClaims(ctx, service_auth.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	err = gp.RemoveMember(ctx, groupID, userID)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Removed user %v from group %v", args[1], args[0]))
	return nil
}
