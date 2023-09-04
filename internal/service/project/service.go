package project

import (
	"context"
	"reflect"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/project"
	project_settings "github.com/rigdev/rig-go-api/api/v1/project/settings"
	storage_settings "github.com/rigdev/rig-go-api/api/v1/storage/settings"
	user_settings "github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/registry"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/internal/gateway/email"
	"github.com/rigdev/rig/internal/oauth2"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/telemetry"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	CreateProject(ctx context.Context, initializers []*project.Update) (*project.Project, error)
	GetProject(ctx context.Context) (*project.Project, error)
	UpdateProject(ctx context.Context, projectID uuid.UUID, us []*project.Update) error
	List(ctx context.Context, query *model.Pagination) (iterator.Iterator[*project.Project], int64, error)
	DeleteProject(ctx context.Context) error

	GetProjectDockerSecret(ctx context.Context, host string) (*registry.Secret, error)

	GetEmailProvider(ctx context.Context) (email.Gateway, error)

	GetSettings(ctx context.Context, settingsType SettingsType, ss proto.Message) error
	SetSettings(ctx context.Context, settingsType SettingsType, ss proto.Message) error
	GetProjectSettings(ctx context.Context) (*project_settings.Settings, error)
	UpdateProjectSettings(ctx context.Context, su []*project_settings.Update) error
}

type service struct {
	cfg config.Config
	rp  repository.Project
	ru  repository.User
	rs  repository.Secret
	ops *oauth2.Providers
}

func NewService(cfg config.Config, rp repository.Project, ru repository.User, rs repository.Secret, ops *oauth2.Providers, t *telemetry.Telemetry) (Service, error) {
	s := &service{
		cfg: cfg,
		rp:  rp,
		ru:  ru,
		rs:  rs,
		ops: ops,
	}

	installationID, err := s.ensureRigProjectExists()
	if err != nil {
		return nil, err
	}

	t.SetInstallationID(installationID)

	return s, nil
}

func (s *service) CreateProject(ctx context.Context, initializers []*project.Update) (*project.Project, error) {
	projectID := uuid.New()

	// Set some some defaults.
	p := DefaultProject()
	p.ProjectId = projectID.String()
	ps := DefaultProjectSettings()
	us := DefaultUserSettings()
	ss := DefaultStorageSettings()

	// TODO: We want to add email/phone if/when a provider is set up.
	for _, i := range initializers {
		if err := applyUpdate(p, i); err != nil {
			return nil, err
		}
	}

	if s.cfg.OAuth.Google.ClientID != "" && s.cfg.OAuth.Google.ClientSecret != "" {
		us.GetOauthSettings().GetGoogle().ClientId = s.cfg.OAuth.Google.ClientID
		us.GetOauthSettings().GetGoogle().AllowLogin = true
		us.GetOauthSettings().GetGoogle().AllowRegister = true
		creds := &model.ProviderCredentials{
			PublicKey:  s.cfg.OAuth.Google.ClientID,
			PrivateKey: s.cfg.OAuth.Google.ClientSecret,
		}

		bytes, err := proto.Marshal(creds)
		if err != nil {
			return nil, err
		}

		secretId := uuid.New()
		err = s.rs.Create(ctx, secretId, bytes)
		if err != nil {
			return nil, err
		}

		us.GetOauthSettings().GetGoogle().SecretId = secretId.String()
	}

	if s.cfg.OAuth.Github.ClientID != "" && s.cfg.OAuth.Github.ClientSecret != "" {
		us.GetOauthSettings().GetGithub().ClientId = s.cfg.OAuth.Github.ClientID
		us.GetOauthSettings().GetGithub().AllowLogin = true
		us.GetOauthSettings().GetGithub().AllowRegister = true

		creds := &model.ProviderCredentials{
			PublicKey:  s.cfg.OAuth.Github.ClientID,
			PrivateKey: s.cfg.OAuth.Github.ClientSecret,
		}
		bytes, err := proto.Marshal(creds)
		if err != nil {
			return nil, err
		}

		secretId := uuid.New()
		err = s.rs.Create(ctx, secretId, bytes)
		if err != nil {
			return nil, err
		}

		us.GetOauthSettings().GetGithub().SecretId = secretId.String()
	}

	if s.cfg.OAuth.Facebook.ClientID != "" && s.cfg.OAuth.Facebook.ClientSecret != "" {
		us.GetOauthSettings().GetFacebook().ClientId = s.cfg.OAuth.Facebook.ClientID
		us.GetOauthSettings().GetFacebook().AllowLogin = true
		us.GetOauthSettings().GetFacebook().AllowRegister = true

		creds := &model.ProviderCredentials{
			PublicKey:  s.cfg.OAuth.Facebook.ClientID,
			PrivateKey: s.cfg.OAuth.Facebook.ClientSecret,
		}
		bytes, err := proto.Marshal(creds)
		if err != nil {
			return nil, err
		}

		secretId := uuid.New()
		err = s.rs.Create(ctx, secretId, bytes)
		if err != nil {
			return nil, err
		}

		us.GetOauthSettings().GetFacebook().SecretId = secretId.String()

		err = s.rs.Create(ctx, secretId, bytes)
		if err != nil {
			return nil, err
		}
	}

	u, err := s.rp.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	// Continue in the context of the new project.
	ctx = auth.WithProjectID(ctx, projectID)

	err = s.SetSettings(ctx, SettingsTypeProject, ps)
	if err != nil {
		return nil, err
	}

	err = s.SetSettings(ctx, SettingsTypeUsers, us)
	if err != nil {
		return nil, err
	}
	err = s.SetSettings(ctx, SettingsTypeStorage, ss)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *service) List(ctx context.Context, query *model.Pagination) (iterator.Iterator[*project.Project], int64, error) {
	// Filter out the Rig project.
	return s.rp.List(ctx, query, []uuid.UUID{auth.RigProjectID})
}

func (s *service) GetProject(ctx context.Context) (*project.Project, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	return s.rp.Get(ctx, projectID)
}

func (s *service) DeleteProject(ctx context.Context) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	if _, err := s.rp.Get(ctx, projectID); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	// TODO: Think about this. How do we clear the databases of legacy content once a project is deleted?
	// if err := s.ru.DeleteAll(ctx); err != nil {
	// 	return err
	// }

	return s.rp.Delete(ctx, projectID)
}

func (s *service) UpdateProject(ctx context.Context, projectID uuid.UUID, us []*project.Update) error {
	p, err := s.rp.Get(ctx, projectID)
	if err != nil {
		return err
	}

	for _, i := range us {
		if err := applyUpdate(p, i); err != nil {
			return err
		}
	}

	if _, err := s.rp.Update(ctx, p); err != nil {
		return err
	}

	return nil
}

func applyUpdate(p *project.Project, pp *project.Update) error {
	switch v := pp.GetField().(type) {
	case *project.Update_Name:
		p.Name = v.Name
		return nil
	default:
		return errors.InvalidArgumentErrorf("invalid user update type '%v'", reflect.TypeOf(v))
	}
}

func DefaultProject() *project.Project {
	return &project.Project{
		CreatedAt: timestamppb.Now(),
	}
}

func DefaultProjectSettings() *project_settings.Settings {
	return &project_settings.Settings{
		EmailProvider: &project_settings.EmailProviderEntry{
			Instance: &project_settings.EmailInstance{
				Instance: &project_settings.EmailInstance_Default{},
			},
		},
		TextProvider: &project_settings.TextProviderEntry{
			Instance: &project_settings.TextInstance{
				Instance: &project_settings.TextInstance_Default{},
			},
		},
		Templates: &project_settings.Templates{
			WelcomeEmail: &project_settings.Template{
				Subject: "Welcome to Rig!",
				Body:    "<h2>Welcome to Rig!</h2><p>Login to get started</p>",
				Type:    project_settings.TemplateType_TEMPLATE_TYPE_WELCOME_EMAIL,
				Format:  []string{"{{ .Identifier }}", "{{ .Email }}"},
			},
			ResetPasswordEmail: &project_settings.Template{
				Subject: "Reset your password",
				Body:    `<h2>Reset your password</h2><p> Enter the following code to reset your password: {{ .Code }}</p>`,
				Type:    project_settings.TemplateType_TEMPLATE_TYPE_EMAIL_RESET_PASSWORD,
				Format:  []string{"{{ .Identifier }}", "{{ .Email }}", "{{ .Code }}"},
			},
			VerifyEmail: &project_settings.Template{
				Subject: "Verify your email",
				Body:    `<h2>Verify your email</h2><p> Enter the following code to verify your email: {{ .Code }}</p>`,
				Type:    project_settings.TemplateType_TEMPLATE_TYPE_EMAIL_VERIFICATION,
				Format:  []string{"{{ .Identifier }}", "{{ .Email }}", "{{ .Code }}"},
			},
		},
	}
}

func DefaultStorageSettings() *storage_settings.Settings {
	return &storage_settings.Settings{}
}

func DefaultUserSettings() *user_settings.Settings {
	return &user_settings.Settings{
		AllowRegister:           true,
		IsVerifiedEmailRequired: true,
		SendWelcomeMail:         true,
		AccessTokenTtl:          durationpb.New(1 * time.Hour),
		RefreshTokenTtl:         durationpb.New(7 * 24 * time.Hour),
		VerificationCodeTtl:     durationpb.New(10 * time.Minute),
		PasswordHashing: &model.HashingConfig{
			Method: &model.HashingConfig_Bcrypt{
				Bcrypt: hash.DefaultBcrypt,
			},
		},
		LoginMechanisms: []model.LoginType{
			model.LoginType_LOGIN_TYPE_USERNAME_PASSWORD,
			model.LoginType_LOGIN_TYPE_EMAIL_PASSWORD,
		},
		OauthSettings: &user_settings.OauthSettings{
			Google: &user_settings.OauthProviderSettings{
				Issuer:        "https://accounts.google.com",
				AllowLogin:    false,
				AllowRegister: false,
			},
			Github: &user_settings.OauthProviderSettings{
				Issuer:        "https://github.com",
				AllowLogin:    false,
				AllowRegister: false,
			},
			Facebook: &user_settings.OauthProviderSettings{
				Issuer:        "https://www.facebook.com",
				AllowLogin:    false,
				AllowRegister: false,
			},
		},
	}
}

func (s *service) ensureRigProjectExists() (uuid.UUID, error) {
	ctx := context.Background()

	/*
		TODO:
			- Ensure settings from configfile are updatable
			- Figure out how to configure registration
	*/

	p, err := s.rp.Get(ctx, auth.RigProjectID)
	if errors.IsNotFound(err) {
		p = DefaultProject()
		p.ProjectId = auth.RigProjectID.String()
		p.Name = "Rig"
		p.InstallationId = uuid.New().String()

		if _, err := s.rp.Create(ctx, p); err != nil {
			return uuid.Nil, err
		}

		us := DefaultUserSettings()
		us.AllowRegister = false
		us.IsVerifiedEmailRequired = false
		us.SendWelcomeMail = false

		if s.cfg.PublicURL != "" {
			us.GetOauthSettings().CallbackUrls = []string{s.cfg.PublicURL}
		}

		if s.cfg.OAuth.Google.ClientID != "" && s.cfg.OAuth.Google.ClientSecret != "" {
			us.GetOauthSettings().GetGoogle().ClientId = s.cfg.OAuth.Google.ClientID
			us.GetOauthSettings().GetGoogle().AllowLogin = true
			us.GetOauthSettings().GetGoogle().AllowRegister = true
			creds := &model.ProviderCredentials{
				PublicKey:  s.cfg.OAuth.Google.ClientID,
				PrivateKey: s.cfg.OAuth.Google.ClientSecret,
			}

			bytes, err := proto.Marshal(creds)
			if err != nil {
				return uuid.Nil, err
			}

			secretId := uuid.New()
			err = s.rs.Create(ctx, secretId, bytes)
			if err != nil {
				return uuid.Nil, err
			}

			us.GetOauthSettings().GetGoogle().SecretId = secretId.String()
		}

		if s.cfg.OAuth.Github.ClientID != "" && s.cfg.OAuth.Github.ClientSecret != "" {
			us.GetOauthSettings().GetGithub().ClientId = s.cfg.OAuth.Github.ClientID
			us.GetOauthSettings().GetGithub().AllowLogin = true
			us.GetOauthSettings().GetGithub().AllowRegister = true

			creds := &model.ProviderCredentials{
				PublicKey:  s.cfg.OAuth.Github.ClientID,
				PrivateKey: s.cfg.OAuth.Github.ClientSecret,
			}
			bytes, err := proto.Marshal(creds)
			if err != nil {
				return uuid.Nil, err
			}

			secretId := uuid.New()
			err = s.rs.Create(ctx, secretId, bytes)
			if err != nil {
				return uuid.Nil, err
			}

			us.GetOauthSettings().GetGithub().SecretId = secretId.String()
		}

		if s.cfg.OAuth.Facebook.ClientID != "" && s.cfg.OAuth.Facebook.ClientSecret != "" {
			us.GetOauthSettings().GetFacebook().ClientId = s.cfg.OAuth.Facebook.ClientID
			us.GetOauthSettings().GetFacebook().AllowLogin = true
			us.GetOauthSettings().GetFacebook().AllowRegister = true

			creds := &model.ProviderCredentials{
				PublicKey:  s.cfg.OAuth.Facebook.ClientID,
				PrivateKey: s.cfg.OAuth.Facebook.ClientSecret,
			}
			bytes, err := proto.Marshal(creds)
			if err != nil {
				return uuid.Nil, err
			}

			secretId := uuid.New()
			err = s.rs.Create(ctx, secretId, bytes)
			if err != nil {
				return uuid.Nil, err
			}

			us.GetOauthSettings().GetFacebook().SecretId = secretId.String()

			err = s.rs.Create(ctx, secretId, bytes)
			if err != nil {
				return uuid.Nil, err
			}
		}

		ss := DefaultStorageSettings()
		ps := DefaultProjectSettings()

		ctx = auth.WithProjectID(ctx, auth.RigProjectID)
		if err := s.SetSettings(ctx, SettingsTypeProject, ps); err != nil {
			return uuid.Nil, err
		}
		if err := s.SetSettings(ctx, SettingsTypeUsers, us); err != nil {
			return uuid.Nil, err
		}
		if err := s.SetSettings(ctx, SettingsTypeStorage, ss); err != nil {
			return uuid.Nil, err
		}
	} else if err != nil {
		return uuid.Nil, err
	}

	installationId := uuid.UUID(p.GetInstallationId())

	if installationId.IsNil() {
		p.InstallationId = uuid.New().String()
		if _, err := s.rp.Update(ctx, p); err != nil {
			return uuid.Nil, err
		}
	}

	return installationId, nil
}
