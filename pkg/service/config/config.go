package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type Service interface {
	Get() *v1alpha1.OperatorConfig
}

func NewService(path string, scheme *runtime.Scheme) (Service, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	cfg, err := deserialize(bs, scheme)
	if err != nil {
		return nil, err
	}
	cfg.Default()

	return &service{cfg: cfg}, nil
}

type service struct {
	cfg *v1alpha1.OperatorConfig
}

// Get implements Service.
func (s *service) Get() *v1alpha1.OperatorConfig {
	return s.cfg
}

func deserialize(data []byte, scheme *runtime.Scheme) (*v1alpha1.OperatorConfig, error) {
	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()
	_, gvk, err := decoder.Decode(data, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decode config: %w", err)
	}

	if gvk.Group != v1alpha1.GroupVersion.Group {
		return nil, errors.New("unsupported api group")
	}

	if gvk.Kind != "OperatorConfig" {
		return nil, errors.New("unsupported api kind")
	}

	cfg := &v1alpha1.OperatorConfig{}

	switch gvk.Version {
	case v1alpha1.GroupVersion.Version:
		if _, _, err := decoder.Decode(data, nil, cfg); err != nil {
			return nil, fmt.Errorf("could not decode into kind: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported api version: %w", err)
	}

	return cfg, nil
}
