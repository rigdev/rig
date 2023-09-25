package metrics

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service interface {
	List(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.InstanceMetrics], error)
	ListWhereCapsuleID(ctx context.Context, pagination *model.Pagination, capsuleID string) (iterator.Iterator[*capsule.InstanceMetrics], error)
	ListWhereCapsuleAndInstanceID(ctx context.Context, pagination *model.Pagination, capsuleID string, instanceID string) (iterator.Iterator[*capsule.InstanceMetrics], error)
}

type service struct {
	cancel  context.CancelFunc
	log     *zap.Logger
	cluster cluster.Gateway
	project project.Service
	cr      repository.Capsule
}

func (s *service) List(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.InstanceMetrics], error) {
	return s.cr.ListMetrics(ctx, pagination)
}

func (s *service) ListWhereCapsuleID(ctx context.Context, pagination *model.Pagination, capsuleID string) (iterator.Iterator[*capsule.InstanceMetrics], error) {
	return s.cr.GetMetrics(ctx, pagination, capsuleID)
}

func (s *service) ListWhereCapsuleAndInstanceID(ctx context.Context, pagination *model.Pagination, capsuleID string, instanceID string) (iterator.Iterator[*capsule.InstanceMetrics], error) {
	return s.cr.GetInstanceMetrics(ctx, pagination, capsuleID, instanceID)
}

type NewParams struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger            *zap.Logger
	Cluster           cluster.Gateway
	Project           project.Service
	CapsuleRepository repository.Capsule
}

const (
	metricResolution = 10 * time.Second
	metricDuration   = 15 * time.Minute // TODO: should be in config
)

func NewService(p NewParams) Service {
	s := &service{
		log:     p.Logger,
		cluster: p.Cluster,
		project: p.Project,
		cr:      p.CapsuleRepository,
	}

	p.Lifecycle.Append(fx.StartStopHook(s.start, s.stop))

	return s
}

func (s *service) start() {
	t := time.NewTicker(metricResolution)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		s.log.Info("metric service started")
		for {
			select {
			case <-ctx.Done():
				t.Stop()
				s.log.Info("metric service stopped", zap.Error(ctx.Err()))
				return
			case <-t.C:
				s.log.Debug("metric update starting")
				if err := s.update(ctx); err != nil {
					s.log.Error("could not update metrics", zap.Error(err))
					continue
				}
				s.log.Debug("metric update finished")
			}
		}
	}()
}

func (s *service) stop() {
	s.cancel()
}

func (s *service) update(ctx context.Context) error {
	var pids []string
	p := &model.Pagination{
		Limit: 100,
	}
	for {
		it, total, err := s.project.List(ctx, p)
		if err != nil {
			return err
		}
		for {
			e, err := it.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			pids = append(pids, e.GetProjectId())
		}

		p.Offset = p.Offset + p.Limit
		if p.Offset > uint32(total) {
			break
		}
	}

	for _, pid := range pids {
		projectCtx := auth.WithProjectID(ctx, pid)
		iter, err := s.cluster.ListCapsuleMetrics(projectCtx)
		if err != nil {
			s.log.Info("failed to read metrics for project", zap.String("project_id", pid), zap.Error(err))
			continue
		}

		for {
			cms, err := iter.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				s.log.Info("failed to read metrics for project", zap.String("project_id", pid), zap.Error(err))
				break
			}

			if err := s.cr.CreateMetrics(projectCtx, cms); err != nil {
				s.log.Info("failed to write metrics for project", zap.String("project_id", pid), zap.Error(err))
				break
			}

			iter.Close()
		}
	}
	return nil
}
