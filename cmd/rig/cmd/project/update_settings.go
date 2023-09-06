package project

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	settingsField      int32
	templateField      int32
	emailProviderField int32
)

const (
	settingsUndefined settingsField = iota
	settingsEmailProvider
	settingsAddDockerRegistry
	settingsDeleteDockerRegistry
	templateEmailWelcome
	templateVerifyEmail
	templateResetPasswordEmail
)

const (
	templateFieldUndefined templateField = iota
	tempalteFieldSubject
	templateFieldBody
)

const (
	emailProviderFieldUndefined emailProviderField = iota
	emailProviderPublicKey
	emailProviderPrivateKey
	emailProviderFromEmail
	emailProviderHost
	emailProviderPort
)

func (f templateField) String() string {
	switch f {
	case tempalteFieldSubject:
		return "Subject"
	case templateFieldBody:
		return "Body"
	default:
		return "Undefined"
	}
}

func (f emailProviderField) String() string {
	switch f {
	case emailProviderPublicKey:
		return "Public Key"
	case emailProviderPrivateKey:
		return "Private Key"
	case emailProviderFromEmail:
		return "From Email"
	case emailProviderHost:
		return "Host"
	case emailProviderPort:
		return "Port"
	default:
		return "Undefined"
	}
}

func (f settingsField) String() string {
	switch f {
	case settingsEmailProvider:
		return "Email Provider"
	case settingsAddDockerRegistry:
		return "Add Docker Registry"
	case settingsDeleteDockerRegistry:
		return "Delete Docker Registry"
	case templateEmailWelcome:
		return "Welcome Email Template"
	case templateVerifyEmail:
		return "Verify Email Template"
	case templateResetPasswordEmail:
		return "Reset Password Email Template"
	default:
		return "Undefined"
	}
}

func ProjectUpdateSettings(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	res, err := nc.ProjectSettings().GetSettings(ctx, &connect.Request[settings.GetSettingsRequest]{})
	if err != nil {
		return err
	}

	s := res.Msg.GetSettings()
	updates := []*settings.Update{}

	if field != "" && value != "" {
		u, err := parseSettingsUpdate()
		if err != nil {
			return err
		}

		_, err = nc.ProjectSettings().UpdateSettings(ctx, &connect.Request[settings.UpdateSettingsRequest]{
			Msg: &settings.UpdateSettingsRequest{
				Updates: []*settings.Update{u},
			},
		})
		if err != nil {
			return err
		}

		cmd.Println("Project settings updated")
		return nil
	}

	fields := []string{
		settingsEmailProvider.String(),
		settingsAddDockerRegistry.String(),
		settingsDeleteDockerRegistry.String(),
		templateEmailWelcome.String(),
		templateVerifyEmail.String(),
		templateResetPasswordEmail.String(),
		"Done",
	}

	for {
		i, res, err := common.PromptSelect("Choose a field to update:", fields, true)
		if err != nil {
			return err
		}
		if res == "Done" {
			break
		}
		u, err := promptSettingsUpdate(settingsField(i+1), s)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if u != nil {
			updates = append(updates, u)
		}
	}

	if len(updates) == 0 {
		cmd.Println("No settings updated")
		return nil
	}

	_, err = nc.ProjectSettings().UpdateSettings(ctx, &connect.Request[settings.UpdateSettingsRequest]{
		Msg: &settings.UpdateSettingsRequest{
			Updates: updates,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Users settings updated")

	return nil
}

func promptSettingsUpdate(f settingsField, s *settings.Settings) (*settings.Update, error) {
	switch f {
	case settingsEmailProvider:
		return promptEmailProvider(s)
	case settingsAddDockerRegistry:
		return promptAddDockerRegistry(s)
	case settingsDeleteDockerRegistry:
		return promptDeleteDockerRegistry(s)
	case templateEmailWelcome:
		return promptTemplate(s.GetTemplates().GetWelcomeEmail())
	case templateResetPasswordEmail:
		return promptTemplate(s.GetTemplates().GetResetPasswordEmail())
	case templateVerifyEmail:
		return promptTemplate(s.GetTemplates().GetVerifyEmail())
	default:
		return nil, nil
	}
}

func promptEmailProvider(s *settings.Settings) (*settings.Update, error) {
	_, field, err := common.PromptSelect("Choose a type:", []string{
		"MailJet",
		"Smtp",
		"Default",
	}, false)
	if err != nil {
		return nil, nil
	}

	switch field {
	case "Default":
		prov := &settings.EmailProvider{
			Instance: &settings.EmailInstance{
				Instance: &settings.EmailInstance_Default{
					Default: &settings.DefaultInstance{},
				},
			},
		}
		return &settings.Update{
			Field: &settings.Update_EmailProvider{
				EmailProvider: prov,
			},
		}, nil
	case "MailJet":
		prov := &settings.EmailProvider{
			Instance:    s.GetEmailProvider().GetInstance(),
			From:        s.GetEmailProvider().GetFrom(),
			Credentials: &model.ProviderCredentials{},
		}
		if prov.GetInstance() == nil || prov.GetInstance().GetMailjet() == nil {
			prov.GetInstance().Instance = &settings.EmailInstance_Mailjet{
				Mailjet: &settings.MailjetInstance{},
			}
		}
		promptEmailProviderFields(prov, field)
		return &settings.Update{
			Field: &settings.Update_EmailProvider{
				EmailProvider: prov,
			},
		}, nil

	case "Smtp":
		prov := &settings.EmailProvider{
			Instance:    s.GetEmailProvider().GetInstance(),
			From:        s.GetEmailProvider().GetFrom(),
			Credentials: &model.ProviderCredentials{},
		}
		if prov.GetInstance() == nil || prov.GetInstance().GetSmtp() == nil {
			prov.GetInstance().Instance = &settings.EmailInstance_Smtp{
				Smtp: &settings.SmtpInstance{},
			}
		}
		promptEmailProviderFields(prov, field)
		return &settings.Update{
			Field: &settings.Update_EmailProvider{
				EmailProvider: prov,
			},
		}, nil
	default:
		return nil, nil
	}
}

func promptDeleteDockerRegistry(s *settings.Settings) (*settings.Update, error) {
	if len(s.GetDockerRegistries()) == 0 {
		return nil, nil
	}

	var hosts []string
	for _, r := range s.GetDockerRegistries() {
		hosts = append(hosts, r.GetHost())
	}

	_, res, err := common.PromptSelect("Choose a registry to delete:", hosts, false)
	if err != nil {
		return nil, err
	}

	return &settings.Update{
		Field: &settings.Update_DeleteDockerRegistry{
			DeleteDockerRegistry: res,
		},
	}, nil
}

func promptAddDockerRegistry(s *settings.Settings) (*settings.Update, error) {
	host, err := common.PromptGetInput("Enter host", common.ValidateNonEmpty)
	if err != nil {
		return nil, err
	}

	username, err := common.PromptGetInput("Enter username", common.ValidateNonEmpty)
	if err != nil {
		return nil, err
	}

	password, err := common.PromptGetInput("Enter password", common.ValidateNonEmpty)
	if err != nil {
		return nil, err
	}

	email, err := common.PromptGetInput("Enter email", common.ValidateEmail)
	if err != nil {
		return nil, err
	}

	reg := &settings.AddDockerRegistry{
		Host: host,
		Field: &settings.AddDockerRegistry_Credentials{
			Credentials: &settings.DockerRegistryCredentials{
				Username: username,
				Password: password,
				Email:    email,
			},
		},
	}
	return &settings.Update{
		Field: &settings.Update_AddDockerRegistry{
			AddDockerRegistry: reg,
		},
	}, nil
}

func promptEmailProviderFields(p *settings.EmailProvider, prov string) error {
	var fields []string
	if prov == "MailJet" {
		fields = []string{
			emailProviderPublicKey.String(),
			emailProviderPrivateKey.String(),
			emailProviderFromEmail.String(),
			"Done",
		}
	} else if prov == "Smtp" {
		fields = []string{
			emailProviderPublicKey.String(),
			emailProviderPrivateKey.String(),
			emailProviderFromEmail.String(),
			emailProviderHost.String(),
			emailProviderPort.String(),
			"Done",
		}
	}

	for {
		_, res, err := common.PromptSelect("Choose a field to update:", fields, true)
		if err != nil {
			return err
		}
		if res == "Done" {
			break
		}

		switch res {
		case emailProviderPublicKey.String():
			key, err := common.PromptGetInput("Enter public key", common.ValidateNonEmpty)
			if err != nil {
				return err
			}
			p.Credentials.PublicKey = key
		case emailProviderPrivateKey.String():
			key, err := common.PromptGetInput("Enter private key", common.ValidateNonEmpty)
			if err != nil {
				return err
			}
			p.Credentials.PrivateKey = key
		case emailProviderFromEmail.String():
			email, err := common.PromptGetInputWithDefault("Enter from email", common.ValidateEmail, p.GetFrom())
			if err != nil {
				return err
			}
			p.From = email
		case emailProviderHost.String():
			host, err := common.PromptGetInputWithDefault("Enter host", common.ValidateNonEmpty, p.GetInstance().GetSmtp().GetHost())
			if err != nil {
				return err
			}
			p.GetInstance().GetSmtp().Host = host
		case emailProviderPort.String():
			port, err := common.PromptGetInputWithDefault("Enter port", common.ValidateNonEmpty, strconv.Itoa(int(p.GetInstance().GetSmtp().GetPort())))
			if err != nil {
				return err
			}
			// parse port as int64
			portInt, err := strconv.Atoi(port)
			if err != nil {
				return err
			}
			p.GetInstance().GetSmtp().Port = int64(portInt)
		default:
			return nil
		}
	}
	return nil
}

func promptTemplate(t *settings.Template) (*settings.Update, error) {
	fields := []string{
		tempalteFieldSubject.String(),
		templateFieldBody.String(),
		"Done",
	}

	for {
		_, res, err := common.PromptSelect("Choose a field to update:", fields, true)
		if err != nil {
			return nil, err
		}
		if res == "Done" {
			break
		}

		switch res {
		case tempalteFieldSubject.String():
			subject, err := common.PromptGetInputWithDefault("Enter subject", common.ValidateNonEmpty, t.GetSubject())
			if err != nil {
				return nil, err
			}
			t.Subject = subject
		case templateFieldBody.String():
			body, err := common.PromptGetInputWithDefault("Enter body", common.ValidateNonEmpty, t.GetBody())
			if err != nil {
				return nil, err
			}
			t.Body = body
		}
	}
	return &settings.Update{
		Field: &settings.Update_Template{
			Template: t,
		},
	}, nil
}

func parseSettingsUpdate() (*settings.Update, error) {
	switch field {
	case common.FormatField(settingsEmailProvider.String()):
		jsonValue := []byte(value)
		prov := settings.EmailProvider{}
		if err := protojson.Unmarshal(jsonValue, &prov); err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_EmailProvider{
				EmailProvider: &prov,
			},
		}, nil
	case common.FormatField(settingsAddDockerRegistry.String()):
		jsonValue := []byte(value)
		reg := settings.AddDockerRegistry{}
		if err := protojson.Unmarshal(jsonValue, &reg); err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_AddDockerRegistry{
				AddDockerRegistry: &reg,
			},
		}, nil
	case common.FormatField(settingsDeleteDockerRegistry.String()):
		return &settings.Update{
			Field: &settings.Update_DeleteDockerRegistry{
				DeleteDockerRegistry: value,
			},
		}, nil
	case "template":
		jsonValue := []byte(value)
		t := settings.Template{}
		if err := protojson.Unmarshal(jsonValue, &t); err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_Template{
				Template: &t,
			},
		}, nil
	default:
		return nil, nil
	}
}
