package project

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	project_settings "github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *service) GetProjectDockerSecret(ctx context.Context) ([]byte, error) {
	ps, err := s.GetProjectSettings(ctx)
	if err != nil {
		return nil, err
	}

	stringSid := ps.DockerRegistries.GetSecretId()
	if stringSid == "" {
		return nil, nil
	}

	bs, err := s.rs.Get(ctx, uuid.UUID(stringSid))
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (s *service) applyAddDockerRegistry(
	ctx context.Context,
	set *project_settings.Settings,
	reg *project_settings.Update_AddDockerRegistry,
) error {
	cfg := createDockerConfigJSON(reg)
	if set.DockerRegistries == nil {
		set.DockerRegistries = &project_settings.DockerRegistries{}
	}

	if sid := set.DockerRegistries.GetSecretId(); sid != "" {
		id := uuid.UUID(sid)

		existingCFG, err := s.getDockerSecret(ctx, id)
		if err != nil {
			return err
		}

		existingCFG.merge(cfg)
		if err := s.updateDockerSecret(ctx, existingCFG, id); err != nil {
			return err
		}
	} else {
		id, err := s.createDockerSecret(ctx, cfg)
		if err != nil {
			return err
		}
		set.DockerRegistries.SecretId = id.String()
	}

	found := false
	for _, h := range set.DockerRegistries.GetHosts() {
		if h == reg.AddDockerRegistry.GetHost() {
			found = true
		}
	}
	if !found {
		set.DockerRegistries.Hosts = append(set.DockerRegistries.Hosts, reg.AddDockerRegistry.GetHost())
	}

	return nil
}

func (s *service) applyDeleteDockerRegistry(
	ctx context.Context,
	set *project_settings.Settings,
	del *project_settings.Update_DeleteDockerRegistry,
) error {
	if set.DockerRegistries == nil {
		return nil
	}
	sid := set.DockerRegistries.GetSecretId()
	if sid == "" {
		return nil
	}

	id := uuid.UUID(sid)

	existingCFG, err := s.getDockerSecret(ctx, id)
	if err != nil {
		return err
	}
	if _, ok := existingCFG.Auths[del.DeleteDockerRegistry]; ok {
		delete(existingCFG.Auths, del.DeleteDockerRegistry)
		if len(existingCFG.Auths) == 0 {
			if err := s.deleteDockerSecret(ctx, id); err != nil {
				return err
			}
		} else {
			if err := s.updateDockerSecret(ctx, existingCFG, id); err != nil {
				return err
			}
		}
	}

	var hs []string
	for _, h := range set.DockerRegistries.GetHosts() {
		if h == del.DeleteDockerRegistry {
			continue
		}
		hs = append(hs, h)
	}
	if len(hs) == 0 {
		set.DockerRegistries = nil
	} else {
		set.DockerRegistries.Hosts = hs
	}

	return nil
}

func (s *service) createDockerSecret(ctx context.Context, reg *dockerConfigJSON) (*uuid.UUID, error) {
	bs, err := json.Marshal(reg)
	if err != nil {
		return nil, fmt.Errorf("could not marshal docker config: %w", err)
	}

	id := uuid.New()

	if err := s.rs.Create(ctx, id, bs); err != nil {
		return nil, fmt.Errorf("could not create docker config.json: %w", err)
	}

	return &id, nil
}

func (s *service) getDockerSecret(ctx context.Context, id uuid.UUID) (*dockerConfigJSON, error) {
	bs, err := s.rs.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	var cfg dockerConfigJSON
	if err := json.NewDecoder(bytes.NewReader(bs)).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("could not decode docker secret: %w", err)
	}
	return &cfg, nil
}

func (s *service) updateDockerSecret(ctx context.Context, cfg *dockerConfigJSON, id uuid.UUID) error {
	bs, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("could not marshal docker config: %w", err)
	}

	if err := s.rs.Update(ctx, id, bs); err != nil {
		return fmt.Errorf("could not update docker config.json: %w", err)
	}

	return nil
}

func (s *service) deleteDockerSecret(ctx context.Context, id uuid.UUID) error {
	if err := s.rs.Delete(ctx, id); err != nil {
		return fmt.Errorf("could not delete docker config.json secret: %w", err)
	}
	return nil
}

type dockerConfigJSON struct {
	Auths map[string]dockerConfigJSONAuth `json:"auths"`
}

func createDockerConfigJSON(reg *project_settings.Update_AddDockerRegistry) *dockerConfigJSON {
	cfg := dockerConfigJSON{Auths: map[string]dockerConfigJSONAuth{}}
	switch v := reg.AddDockerRegistry.Field.(type) {
	case *project_settings.DockerRegistry_Auth:
		cfg.Auths[reg.AddDockerRegistry.GetHost()] = dockerConfigJSONAuth{
			Auth: reg.AddDockerRegistry.GetAuth(),
		}
	case *project_settings.DockerRegistry_Credentials:
		up := fmt.Sprintf("%s:%s", v.Credentials.GetUsername(), v.Credentials.GetPassword())
		cfg.Auths[reg.AddDockerRegistry.GetHost()] = dockerConfigJSONAuth{
			Auth:     base64.StdEncoding.EncodeToString([]byte(up)),
			Username: v.Credentials.GetUsername(),
			Password: v.Credentials.GetPassword(),
			Email:    v.Credentials.GetEmail(),
		}
	}
	return &cfg
}

func (cfg *dockerConfigJSON) merge(newCFG *dockerConfigJSON) {
	for h, a := range newCFG.Auths {
		cfg.Auths[h] = a
	}
}

type dockerConfigJSONAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Auth     string `json:"auth"`
}
