package capabilities

import (
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
)

type Service interface {
	Get() (*capabilities.GetResponse, error)
}

func NewService(cfg config.Service) Service {
	return &service{cfg: cfg}
}

type service struct {
	cfg config.Service
}

// Get implements Service.
func (s *service) Get() (*capabilities.GetResponse, error) {
	res := &capabilities.GetResponse{}

	cfg := s.cfg.Get()

	if cfg.Certmanager != nil && cfg.Certmanager.ClusterIssuer != "" {
		res.Ingress = true
	}

	return res, nil
}
