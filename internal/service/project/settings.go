package project

import (
	"context"
	"errors"

	project_settings "github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

type SettingsType string

const (
	SettingsTypeUsers   SettingsType = "users"
	SettingsTypeStorage SettingsType = "storage"
	SettingsTypeProject SettingsType = "project"
)

func (s *service) SetSettings(ctx context.Context, settingsType SettingsType, ss proto.Message) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	settings, err := proto.Marshal(ss)
	if err != nil {
		return err
	}
	return s.rp.SetSettings(ctx, projectID, string(settingsType), settings)
}

func (s *service) GetSettings(ctx context.Context, settingsType SettingsType, ss proto.Message) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	bs, err := s.rp.GetSettings(ctx, projectID, string(settingsType))
	if err != nil {
		return err
	}

	return proto.Unmarshal(bs, ss)
}

func (s *service) GetProjectSettings(ctx context.Context) (*project_settings.Settings, error) {
	res := &project_settings.Settings{}
	err := s.GetSettings(ctx, SettingsTypeProject, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *service) UpdateProjectSettings(ctx context.Context, su []*project_settings.Update) error {
	set, err := s.GetProjectSettings(ctx)
	if err != nil {
		return err
	}

	err = s.applySettingsUpdates(ctx, set, su)
	if err != nil {
		return err
	}

	return s.SetSettings(ctx, SettingsTypeProject, set)
}

func (s *service) applySettingsUpdates(ctx context.Context, set *project_settings.Settings, us []*project_settings.Update) error {
	for _, u := range us {
		switch v := u.Field.(type) {
		case *project_settings.Update_EmailProvider:
			bytes, err := proto.Marshal(v.EmailProvider.GetCredentials())
			if err != nil {
				return err
			}

			protoSecretID := set.GetEmailProvider().GetSecretId()
			if protoSecretID == "" {
				secretID := uuid.New()

				err = s.rs.Create(ctx, secretID, bytes)
				if err != nil {
					return err
				}

				protoSecretID = secretID.String()
			} else {
				err = s.rs.Update(ctx, uuid.UUID(protoSecretID), bytes)
				if err != nil {
					return err
				}
			}

			entry := &project_settings.EmailProviderEntry{
				From:     v.EmailProvider.GetFrom(),
				ClientId: v.EmailProvider.GetCredentials().GetPublicKey(),
				Instance: v.EmailProvider.GetInstance(),
				SecretId: protoSecretID,
			}

			set.EmailProvider = entry
		case *project_settings.Update_TextProvider:
			bytes, err := proto.Marshal(v.TextProvider.Credentials)
			if err != nil {
				return err
			}

			protoSecretID := set.GetEmailProvider().GetSecretId()
			if protoSecretID == "" {
				secretID := uuid.New()

				err = s.rs.Create(ctx, secretID, bytes)
				if err != nil {
					return err
				}

				protoSecretID = secretID.String()
			} else {
				err = s.rs.Update(ctx, uuid.UUID(protoSecretID), bytes)
				if err != nil {
					return err
				}
			}

			entry := &project_settings.TextProviderEntry{
				From:     v.TextProvider.GetFrom(),
				ClientId: v.TextProvider.GetCredentials().GetPublicKey(),
				Instance: v.TextProvider.GetInstance(),
				SecretId: protoSecretID,
			}

			set.TextProvider = entry
		case *project_settings.Update_Template:
			var t *project_settings.Template
			switch v.Template.GetType() {
			case project_settings.TemplateType_TEMPLATE_TYPE_WELCOME_EMAIL:
				t = set.Templates.WelcomeEmail
			case project_settings.TemplateType_TEMPLATE_TYPE_EMAIL_VERIFICATION:
				t = set.Templates.VerifyEmail
			case project_settings.TemplateType_TEMPLATE_TYPE_EMAIL_RESET_PASSWORD:
				t = set.Templates.ResetPasswordEmail
			default:
				return errors.New("invalid template type")
			}
			t.Subject = v.Template.Subject
			t.Body = v.Template.Body
		case *project_settings.Update_AddDockerRegistry:
			s.applyAddDockerRegistry(ctx, set, v)
		case *project_settings.Update_DeleteDockerRegistry:
			s.applyDeleteDockerRegistry(ctx, set, v)
		}
	}
	return nil
}
