package capsule

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	logger *zap.Logger
	cr     repository.Capsule
	sr     repository.Secret
	cg     cluster.Gateway
	as     *auth.Service
	ps     project.Service
	q      *Queue[Job]
	cfg    config.Config
}

func NewService(cr repository.Capsule, sr repository.Secret, cg cluster.Gateway, as *auth.Service, ps project.Service, cfg config.Config, logger *zap.Logger) *Service {
	s := &Service{
		cr:     cr,
		sr:     sr,
		cg:     cg,
		as:     as,
		ps:     ps,
		q:      NewQueue[Job](),
		cfg:    cfg,
		logger: logger,
	}

	go s.run()

	return s
}

func (s *Service) CreateCapsule(ctx context.Context, name string, is []*capsule.Update) (uuid.UUID, error) {
	capsuleID := uuid.New()

	c := &capsule.Capsule{
		CapsuleId: capsuleID.String(),
		Name:      name,
		CreatedAt: timestamppb.Now(),
	}

	if a, err := s.as.GetAuthor(ctx); err != nil {
		return uuid.Nil, err
	} else {
		c.CreatedBy = a
	}

	if err := applyUpdates(c, is); err != nil {
		return uuid.Nil, err
	}

	if err := s.cr.Create(ctx, c); err != nil {
		return uuid.Nil, err
	}

	return capsuleID, nil
}

func (s *Service) GetCapsule(ctx context.Context, capsuleID uuid.UUID) (*capsule.Capsule, error) {
	return s.cr.Get(ctx, capsuleID)
}

func (s *Service) GetCapsuleByName(ctx context.Context, name string) (*capsule.Capsule, error) {
	return s.cr.GetByName(ctx, name)
}

func (s *Service) Logs(ctx context.Context, capsuleID uuid.UUID, instanceID string, follow bool) (iterator.Iterator[*capsule.Log], error) {
	d, err := s.cr.Get(ctx, capsuleID)
	if err != nil {
		return nil, err
	}

	return s.cg.Logs(ctx, d.GetName(), instanceID, follow)
}

func (s *Service) DeleteCapsule(ctx context.Context, capsuleID uuid.UUID) error {
	d, err := s.cr.Get(ctx, capsuleID)
	if err != nil {
		return err
	}

	if err := s.cg.DeleteCapsule(ctx, d.GetName()); errors.IsNotFound(err) {
		// The capsule didn't exist.
	} else if err != nil {
		return err
	}

	if err := s.cr.Delete(ctx, capsuleID); err != nil {
		return err
	}

	return nil
}

func (s *Service) ListCapsules(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.Capsule], int64, error) {
	return s.cr.List(ctx, pagination)
}

func (s *Service) UpdateCapsule(ctx context.Context, capsuleID uuid.UUID, us []*capsule.Update) error {
	d, err := s.cr.Get(ctx, capsuleID)
	if err != nil {
		return err
	}

	if err := applyUpdates(d, us); err != nil {
		return err
	}

	if err := s.cr.Update(ctx, d); err != nil {
		return err
	}

	return nil
}

func (s *Service) ListInstances(ctx context.Context, capsuleID uuid.UUID, pagination *model.Pagination) (iterator.Iterator[*capsule.Instance], uint64, error) {
	c, err := s.cr.Get(ctx, capsuleID)
	if err != nil {
		return nil, 0, err
	}

	return s.cg.ListInstances(ctx, c.GetName())
}

func (s *Service) RestartInstance(ctx context.Context, capsuleID uuid.UUID, instanceID string) error {
	c, err := s.cr.Get(ctx, capsuleID)
	if err != nil {
		return err
	}

	return s.cg.RestartInstance(ctx, c.GetName(), instanceID)
}

func (s *Service) DeleteBuild(ctx context.Context, capsuleID uuid.UUID, buildID string) error {
	cp, err := s.cr.Get(ctx, capsuleID)
	if err != nil {
		return err
	}

	ro, _, _, err := s.cr.GetRollout(ctx, capsuleID, cp.GetCurrentRollout())
	if err != nil {
		return err
	}

	// If the build we are deleting is the one currently used by the capsule, delete the deployment if it exists.
	if ro.GetBuildId() == buildID {
		return errors.FailedPreconditionErrorf("cannot delete the current build")
	}

	return s.cr.DeleteBuild(ctx, capsuleID, buildID)
}

func (s *Service) Deploy(ctx context.Context, capsuleID uuid.UUID, cs []*capsule.Change) error {
	if _, err := s.newRollout(ctx, capsuleID, cs); err != nil {
		return err
	}

	return nil
}

func (s *Service) ListRollouts(ctx context.Context, capsuleID uuid.UUID, pagination *model.Pagination) (iterator.Iterator[*capsule.Rollout], uint64, error) {
	return s.cr.ListRollouts(ctx, pagination, capsuleID)
}

func applyUpdates(d *capsule.Capsule, us []*capsule.Update) error {
	for range us {
		return errors.InvalidArgumentErrorf("unknown update field type")
	}
	return nil
}
