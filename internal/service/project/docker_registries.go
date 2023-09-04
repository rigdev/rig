package project

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	project_settings "github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig/gen/go/registry"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

func (s *service) GetProjectDockerSecret(ctx context.Context, host string) (*registry.Secret, error) {
	ps, err := s.GetProjectSettings(ctx)
	if err != nil {
		return nil, err
	}

	for _, p := range ps.GetDockerRegistries() {
		if p.GetHost() == host {
			bs, err := s.rs.Get(ctx, uuid.UUID(p.GetSecretId()))
			if err != nil {
				return nil, err
			}

			rs := &registry.Secret{}
			if err := proto.Unmarshal(bs, rs); err != nil {
				return nil, err
			}

			return rs, nil
		}
	}

	return nil, errors.NotFoundErrorf("registry host not found")
}

func (s *service) applyAddDockerRegistry(
	ctx context.Context,
	set *project_settings.Settings,
	reg *project_settings.Update_AddDockerRegistry,
) error {
	for _, h := range set.GetDockerRegistries() {
		if reg.AddDockerRegistry.GetHost() == h.GetHost() {
			return errors.AlreadyExistsErrorf("registry host already exists")
		}
	}

	rs, err := createRegistrySecret(reg)
	if err != nil {
		return err
	}

	id, err := s.createDockerSecret(ctx, rs)
	if err != nil {
		return err
	}

	set.DockerRegistries = append(set.DockerRegistries, &project_settings.DockerRegistry{
		Host:     reg.AddDockerRegistry.GetHost(),
		SecretId: id.String(),
	})

	return nil
}

func (s *service) applyDeleteDockerRegistry(
	ctx context.Context,
	set *project_settings.Settings,
	reg *project_settings.Update_DeleteDockerRegistry,
) error {
	for i, h := range set.GetDockerRegistries() {
		if reg.DeleteDockerRegistry == h.GetHost() {
			secretID, err := uuid.Parse(h.GetSecretId())
			if err != nil {
				return err
			}

			if err := s.deleteDockerSecret(ctx, secretID); err != nil {
				return err
			}

			set.DockerRegistries = append(
				set.GetDockerRegistries()[:i],
				set.GetDockerRegistries()[i+1:]...,
			)
			return nil
		}
	}

	return errors.NotFoundErrorf("registry not found")
}

func (s *service) createDockerSecret(ctx context.Context, rs *registry.Secret) (*uuid.UUID, error) {
	bs, err := proto.Marshal(rs)
	if err != nil {
		return nil, fmt.Errorf("could not marshal docker config: %w", err)
	}

	id := uuid.New()

	if err := s.rs.Create(ctx, id, bs); err != nil {
		return nil, fmt.Errorf("could not create docker config.json: %w", err)
	}

	return &id, nil
}

func (s *service) deleteDockerSecret(ctx context.Context, id uuid.UUID) error {
	return s.rs.Delete(ctx, id)
}

func createRegistrySecret(reg *project_settings.Update_AddDockerRegistry) (*registry.Secret, error) {
	switch v := reg.AddDockerRegistry.Field.(type) {
	case *project_settings.AddDockerRegistry_Auth:
		raw, err := base64.StdEncoding.DecodeString(v.Auth)
		if err != nil {
			return nil, errors.InvalidArgumentErrorf("invalid auth format, expected base64 encoded payload")
		}

		var r struct {
			Username string
			Password string
			Auth     string
		}
		if err := json.Unmarshal(raw, &r); err == nil {
			if r.Username != "" && r.Password != "" {
				return &registry.Secret{
					Username: r.Username,
					Password: r.Password,
				}, nil
			}

			if r.Auth == "" {
				errors.InvalidArgumentErrorf("invalid auth format, expected base64 json token with `username/password` or `auth`")
			}

			raw, err = base64.StdEncoding.DecodeString(r.Auth)
			if err != nil {
				return nil, errors.InvalidArgumentErrorf("invalid auth format, expected base64 encoded payload")
			}
		}

		if idx := bytes.IndexByte(raw, ':'); idx > 0 && idx < len(raw)-1 {
			return &registry.Secret{
				Username: string(raw[0:idx]),
				Password: string(raw[idx+1:]),
			}, nil
		}

		return nil, errors.InvalidArgumentErrorf("invalid auth format, expected json or `username:password` formatted base64 token")

	case *project_settings.AddDockerRegistry_Credentials:
		return &registry.Secret{
			Username: v.Credentials.Username,
			Password: v.Credentials.Password,
		}, nil

	default:
		return nil, errors.InvalidArgumentErrorf("invalid registry auth type '%v'", reflect.TypeOf(v))
	}
}
