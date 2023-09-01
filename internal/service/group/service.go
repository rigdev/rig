package group

import (
	"context"
	"reflect"

	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	rg     repository.Group
	ug     repository.User
	logger *zap.Logger
}

func NewService(rg repository.Group, ug repository.User, logger *zap.Logger) *Service {
	return &Service{
		rg:     rg,
		ug:     ug,
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, initializers []*group.Update) (*group.Group, error) {
	groupID := uuid.New()
	g := &group.Group{
		GroupId:    groupID.String(),
		CreatedAt:  timestamppb.Now(),
		NumMembers: 0,
	}

	for _, i := range initializers {
		if err := applyUpdate(g, i); err != nil {
			return nil, err
		}
	}

	g, err := s.rg.Create(ctx, g)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (s *Service) Update(ctx context.Context, groupID uuid.UUID, us []*group.Update) error {
	g, err := s.rg.Get(ctx, groupID)
	if err != nil {
		return err
	}

	for _, i := range us {
		if err := applyUpdate(g, i); err != nil {
			return err
		}
	}

	if _, err := s.rg.Update(ctx, g); err != nil {
		return err
	}
	return nil
}

func (s *Service) List(ctx context.Context, pagination *model.Pagination, search string) (iterator.Iterator[*group.Group], uint64, error) {
	it, total, err := s.rg.List(ctx, pagination, search)
	if err != nil {
		return nil, 0, err
	}

	mit := iterator.Map(it, func(g *group.Group) (*group.Group, error) {
		numMembers, err := s.rg.CountMembers(ctx, uuid.UUID(g.GetGroupId()))
		if err != nil {
			return nil, err
		}
		g.NumMembers = numMembers
		return g, nil
	})

	return mit, total, nil
}

func (s *Service) Get(ctx context.Context, groupID uuid.UUID) (*group.Group, error) {
	g, err := s.rg.Get(ctx, groupID)
	if err != nil {
		return nil, err
	}
	numMembers, err := s.rg.CountMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	g.NumMembers = numMembers
	return g, nil
}

func (s *Service) GetByName(ctx context.Context, groupName string) (*group.Group, error) {
	g, err := s.rg.GetByName(ctx, groupName)
	if err != nil {
		return nil, err
	}

	numMembers, err := s.rg.CountMembers(ctx, uuid.UUID(g.GetGroupId()))
	if err != nil {
		return nil, err
	}
	g.NumMembers = numMembers
	return g, nil
}

func (s *Service) Delete(ctx context.Context, groupID uuid.UUID) error {
	err := s.rg.Delete(ctx, groupID)
	return err
}

func (s *Service) Count(ctx context.Context) (int64, error) {
	return s.rg.Count(ctx)
}

func (s *Service) AddMembers(ctx context.Context, groupID uuid.UUID, userIDs []uuid.UUID) error {
	// verify that group exists
	_, err := s.rg.Get(ctx, groupID)
	if err != nil {
		s.logger.Debug("group does not exist", zap.Error(err))
		return err
	}
	for _, userID := range userIDs {
		_, err = s.ug.Get(ctx, userID)
		if err != nil {
			s.logger.Debug("user does not exist", zap.Error(err))
			return err
		}
	}

	return s.rg.AddMembers(ctx, userIDs, groupID)
}

func (s *Service) RemoveMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	return s.rg.RemoveMember(ctx, userID, groupID)
}

func (s *Service) RemoveMemberFromAll(ctx context.Context, userID uuid.UUID) error {
	return s.rg.RemoveMemberFromAll(ctx, userID)
}

func (s *Service) ListGroupsForUser(ctx context.Context, userID uuid.UUID, pagination *model.Pagination) (iterator.Iterator[*group.Group], uint64, error) {
	it, total, err := s.rg.ListGroupsForUser(ctx, userID, pagination)
	if err != nil {
		return nil, 0, err
	}

	itm := iterator.Map(it, func(groupID uuid.UUID) (*group.Group, error) {
		return s.rg.Get(ctx, groupID)
	})
	return itm, total, nil
}

func (s *Service) ListMembers(ctx context.Context, groupID uuid.UUID, pagination *model.Pagination) (iterator.Iterator[*model.MemberEntry], uint64, error) {
	it, total, err := s.rg.ListMembers(ctx, groupID, pagination)
	if err != nil {
		return nil, 0, err
	}

	itm := iterator.Map(it, func(userID uuid.UUID) (*model.MemberEntry, error) {
		u, err := s.ug.Get(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &model.MemberEntry{
			User: &model.UserEntry{
				UserId:        userID.String(),
				PrintableName: utils.UserName(u),
				// TODO: Think about this
				Verified:  u.GetIsEmailVerified(),
				CreatedAt: u.GetUserInfo().GetCreatedAt(),
			},
			// TODO: needs member.joinedAt.
			// JoinedAt: timestamppb.New(member.CreatedAt),
		}, nil
	})

	return itm, total, nil
}

func applyUpdate(g *group.Group, gp *group.Update) error {
	switch v := gp.GetField().(type) {
	case *group.Update_Name:
		g.Name = v.Name
	case *group.Update_SetMetadata:
		if g.Metadata == nil {
			g.Metadata = map[string][]byte{}
		}
		g.Metadata[v.SetMetadata.GetKey()] = v.SetMetadata.GetValue()
	case *group.Update_DeleteMetadataKey:
		delete(g.Metadata, v.DeleteMetadataKey)
	default:
		return errors.InvalidArgumentErrorf("invalid group update type '%v'", reflect.TypeOf(v))
	}
	return nil
}
