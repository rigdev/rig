package capsule

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/gateway/cluster"
	"github.com/rigdev/rig/pkg/repository"
	service_auth "github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Service struct {
	logger *zap.Logger
	cr     repository.Capsule
	sr     repository.Secret
	cg     cluster.Gateway
	ccg    cluster.ConfigGateway
	csg    cluster.StatusGateway
	as     *service_auth.Service
	ps     project.Service
	q      *Queue[Job]
	cfg    config.Config
}

func NewService(
	cr repository.Capsule,
	sr repository.Secret,
	cg cluster.Gateway,
	ccg cluster.ConfigGateway,
	csg cluster.StatusGateway,
	as *service_auth.Service,
	ps project.Service,
	cfg config.Config,
	logger *zap.Logger,
) *Service {
	s := &Service{
		cr:     cr,
		sr:     sr,
		cg:     cg,
		ccg:    ccg,
		csg:    csg,
		as:     as,
		ps:     ps,
		q:      NewQueue[Job](),
		cfg:    cfg,
		logger: logger,
	}

	go s.run()

	return s
}

func (s *Service) CreateCapsule(ctx context.Context, name string, is []*capsule.Update) (string, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return "", err
	}

	cfg := &v1alpha1.Capsule{
		TypeMeta: v1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.String(),
			Kind:       "Capsule",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: projectID,
		},
	}
	if err := s.ccg.CreateCapsuleConfig(ctx, cfg); err != nil {
		return "", err
	}

	return name, nil
}

func (s *Service) GetCapsule(ctx context.Context, capsuleID string) (*capsule.Capsule, error) {
	cfg, err := s.ccg.GetCapsuleConfig(ctx, capsuleID)
	if err != nil {
		return nil, err
	}

	return s.toCapsule(ctx, cfg)
}

func (s *Service) toCapsule(ctx context.Context, cfg *v1alpha1.Capsule) (*capsule.Capsule, error) {
	rolloutID, rc, rs, _, err := s.cr.GetCurrentRollout(ctx, cfg.GetName())
	if errors.IsNotFound(err) {
	} else if err != nil {
		return nil, err
	}

	return &capsule.Capsule{
		CapsuleId:      cfg.GetName(),
		CurrentRollout: rolloutID,
		UpdatedAt:      rs.GetStatus().GetUpdatedAt(),
		UpdatedBy:      rc.GetCreatedBy(),
	}, nil
}

func (s *Service) Logs(ctx context.Context, capsuleID string, instanceID string, follow bool) (iterator.Iterator[*capsule.Log], error) {
	d, err := s.GetCapsule(ctx, capsuleID)
	if err != nil {
		return nil, err
	}

	return s.cg.Logs(ctx, d.GetCapsuleId(), instanceID, follow)
}

func (s *Service) DeleteCapsule(ctx context.Context, capsuleID string) error {
	_, err := s.GetCapsule(ctx, capsuleID)
	if err != nil {
		return err
	}

	if err := s.cr.Delete(ctx, capsuleID); err != nil {
		return err
	}

	if err := s.ccg.DeleteCapsuleConfig(ctx, capsuleID); err != nil {
		return err
	}

	return nil
}

func (s *Service) ListCapsules(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.Capsule], int64, error) {
	it, total, err := s.ccg.ListCapsuleConfigs(ctx, pagination)
	if err != nil {
		return nil, 0, err
	}

	it2 := iterator.Map(it, func(cfg *v1alpha1.Capsule) (*capsule.Capsule, error) {
		return s.toCapsule(ctx, cfg)
	})

	return it2, total, nil
}

func (s *Service) UpdateCapsule(ctx context.Context, capsuleID uuid.UUID, us []*capsule.Update) error {
	return errors.UnimplementedErrorf("UpdateCapsule not implemented")
}

func (s *Service) ListInstances(ctx context.Context, capsuleID string, pagination *model.Pagination) (iterator.Iterator[*capsule.Instance], uint64, error) {
	c, err := s.GetCapsule(ctx, capsuleID)
	if err != nil {
		return nil, 0, err
	}

	return s.cg.ListInstances(ctx, c.GetCapsuleId())
}

func (s *Service) RestartInstance(ctx context.Context, capsuleID string, instanceID string) error {
	c, err := s.GetCapsule(ctx, capsuleID)
	if err != nil {
		return err
	}

	return s.cg.RestartInstance(ctx, c.GetCapsuleId(), instanceID)
}

func (s *Service) DeleteBuild(ctx context.Context, capsuleID string, buildID string) error {
	cp, err := s.GetCapsule(ctx, capsuleID)
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

func (s *Service) Deploy(ctx context.Context, capsuleID string, cs []*capsule.Change) (uint64, error) {
	rolloutID, err := s.newRollout(ctx, capsuleID, cs)
	if err != nil {
		return 0, err
	}

	return rolloutID, nil
}

func (s *Service) ListRollouts(ctx context.Context, capsuleID string, pagination *model.Pagination) (iterator.Iterator[*capsule.Rollout], uint64, error) {
	return s.cr.ListRollouts(ctx, pagination, capsuleID)
}

func applyUpdates(d *capsule.Capsule, us []*capsule.Update) error {
	for range us {
		return errors.InvalidArgumentErrorf("unknown update field type")
	}
	return nil
}

func (s *Service) Rollback(ctx context.Context, capsuleID string, rolloutID uint64) (uint64, error) {
	return 0, nil
}
