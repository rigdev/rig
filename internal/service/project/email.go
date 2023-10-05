package project

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/client/mailjet"
	"github.com/rigdev/rig/pkg/client/smtp"
	"github.com/rigdev/rig/internal/gateway/email"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

func (s *service) GetEmailProvider(ctx context.Context) (email.Gateway, error) {
	set, err := s.GetProjectSettings(ctx)
	if err != nil {
		return nil, err
	}

	switch set.GetEmailProvider().GetInstance().GetInstance().(type) {
	case *settings.EmailInstance_Default:
		switch s.cfg.Email.Type {
		case "mailjet":
			return mailjet.New(&settings.EmailProvider{
				From: s.cfg.Client.Mailjet.From,
				Credentials: &model.ProviderCredentials{
					PublicKey:  s.cfg.Client.Mailjet.APIKey,
					PrivateKey: s.cfg.Client.Mailjet.SecretKey,
				},
				Instance: &settings.EmailInstance{
					Instance: &settings.EmailInstance_Default{
						Default: &settings.DefaultInstance{},
					},
				},
			}), nil
		case "smtp":
			return smtp.New(&settings.EmailProvider{
				From: s.cfg.Email.From,
				Credentials: &model.ProviderCredentials{
					PublicKey:  s.cfg.Client.SMTP.Username,
					PrivateKey: s.cfg.Client.SMTP.Password,
				},
				Instance: &settings.EmailInstance{
					Instance: &settings.EmailInstance_Smtp{
						Smtp: &settings.SmtpInstance{
							Host: s.cfg.Client.SMTP.Host,
							Port: int64(s.cfg.Client.SMTP.Port),
						},
					},
				},
			}), nil

		default:
			return nil, errors.NotFoundErrorf("no default email provider configured")
		}
	case *settings.EmailInstance_Mailjet:
		id := uuid.UUID(set.GetEmailProvider().GetSecretId())

		bytes, err := s.rs.Get(ctx, id)
		if err != nil {
			return nil, err
		}

		credentials := model.ProviderCredentials{}
		err = proto.Unmarshal(bytes, &credentials)
		if err != nil {
			return nil, err
		}

		prov := &settings.EmailProvider{
			From:        set.GetEmailProvider().GetFrom(),
			Instance:    set.GetEmailProvider().GetInstance(),
			Credentials: &credentials,
		}
		return mailjet.New(prov), nil
	case *settings.EmailInstance_Smtp:
		id := uuid.UUID(set.GetEmailProvider().GetSecretId())

		bytes, err := s.rs.Get(ctx, id)
		if err != nil {
			return nil, err
		}

		credentials := model.ProviderCredentials{}
		err = proto.Unmarshal(bytes, &credentials)
		if err != nil {
			return nil, err
		}

		prov := &settings.EmailProvider{
			From:        set.GetEmailProvider().GetFrom(),
			Instance:    set.GetEmailProvider().GetInstance(),
			Credentials: &credentials,
		}

		return smtp.New(prov), nil
	default:
		return nil, errors.NotFoundErrorf("no email provider configured")
	}
}
