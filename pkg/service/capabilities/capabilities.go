package capabilities

import (
	"context"

	"github.com/rigdev/rig/gen/go/operator/api/v1/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
)

type Service interface {
	Get(ctx context.Context, req *capabilities.GetRequest) (*capabilities.GetResponse, error)
}

func NewService(cfg config.Service) Service {
	return &service{cfg: cfg}
}

type service struct {
	cfg config.Service
}

// Get implements Service.
func (s *service) Get(ctx context.Context, req *capabilities.GetRequest) (*capabilities.GetResponse, error) {
	res := &capabilities.GetResponse{}

	cfg := s.cfg.Get()

	if cfg.Certmanager != nil && cfg.Certmanager.ClusterIssuer != "" {
		res.Ingress = true
	}

	return res, nil
}
