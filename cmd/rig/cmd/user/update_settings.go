package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
)

func loginTypeToString(l model.LoginType) string {
	switch l {
	case model.LoginType_LOGIN_TYPE_EMAIL_PASSWORD:
		return "Email/Password"
	case model.LoginType_LOGIN_TYPE_PHONE_PASSWORD:
		return "Phone/Password"
	case model.LoginType_LOGIN_TYPE_USERNAME_PASSWORD:
		return "Username/Password"
	default:
		return "Unknown"
	}
}

func loginTypeFromString(s string) model.LoginType {
	switch s {
	case "Email/Password":
		return model.LoginType_LOGIN_TYPE_EMAIL_PASSWORD
	case "Phone/Password":
		return model.LoginType_LOGIN_TYPE_PHONE_PASSWORD
	case "Username/Password":
		return model.LoginType_LOGIN_TYPE_USERNAME_PASSWORD
	default:
		return model.LoginType_LOGIN_TYPE_UNSPECIFIED
	}
}

type (
	settingsField      int32
	templateField      int32
	emailProviderField int32
)

const (
	settingsUndefined settingsField = iota
	settingsAllowRegister
	settingsIsVerifiedEmailRequired
	settingsIsVerifiedPhoneRequired
	settingsAccessTokenTTL
	settingsRefreshTokenTTL
	settingsVerificationCodeTTL
	settingsPasswordHashing
	settingsLoginMechanisms
	settingsEmailProvider
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
	case settingsAllowRegister:
		return "Allow Register"
	case settingsIsVerifiedEmailRequired:
		return "Verify Email Required"
	case settingsIsVerifiedPhoneRequired:
		return "Verify Phone Required"
	case settingsAccessTokenTTL:
		return "Access Token TTL"
	case settingsRefreshTokenTTL:
		return "Refresh Token TTL"
	case settingsVerificationCodeTTL:
		return "Verification Code TTL"
	case settingsPasswordHashing:
		return "Password Hashing"
	case settingsLoginMechanisms:
		return "Login Mechanisms"
	case settingsEmailProvider:
		return "Email Provider"
	case templateEmailWelcome:
		return "Welcome Email Template"
	case templateVerifyEmail:
		return "Verify Email Template"
	case templateResetPasswordEmail:
		return "Reset Password Email Template"
	default:
		return "Unknown"
	}
}

func (c *Cmd) updateSettings(ctx context.Context, cmd *cobra.Command, _ []string) error {
	res, err := c.Rig.UserSettings().GetSettings(ctx, &connect.Request[settings.GetSettingsRequest]{})
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

		_, err = c.Rig.UserSettings().UpdateSettings(ctx, &connect.Request[settings.UpdateSettingsRequest]{
			Msg: &settings.UpdateSettingsRequest{
				Settings: []*settings.Update{u},
			},
		})
		if err != nil {
			return err
		}

		cmd.Println("Users settings updated")
		return nil
	}

	fields := []string{
		settingsAllowRegister.String(),
		settingsIsVerifiedEmailRequired.String(),
		settingsIsVerifiedPhoneRequired.String(),
		settingsAccessTokenTTL.String(),
		settingsRefreshTokenTTL.String(),
		settingsVerificationCodeTTL.String(),
		settingsPasswordHashing.String(),
		settingsLoginMechanisms.String(),
		settingsEmailProvider.String(),
		templateEmailWelcome.String(),
		templateVerifyEmail.String(),
		templateResetPasswordEmail.String(),
		"Done",
	}

	for {
		i, res, err := c.Prompter.Select("Choose a field to update:", fields)
		if err != nil {
			return err
		}
		if res == "Done" {
			break
		}
		u, err := c.promptSettingsUpdate(settingsField(i+1), s)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if u != nil {
			updates = append(updates, u...)
		}
	}

	if len(updates) == 0 {
		cmd.Println("No settings updated")
		return nil
	}

	_, err = c.Rig.UserSettings().UpdateSettings(ctx, &connect.Request[settings.UpdateSettingsRequest]{
		Msg: &settings.UpdateSettingsRequest{
			Settings: updates,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Users settings updated")

	return nil
}

func (c *Cmd) promptSettingsUpdate(f settingsField, s *settings.Settings) ([]*settings.Update, error) {
	switch f {
	case settingsAllowRegister:
		defAllowRegister := strconv.FormatBool(s.GetAllowRegister())
		allowRegister, err := c.Prompter.Input(
			"Allow Register:", common.BoolValidateOpt, common.InputDefaultOpt(defAllowRegister),
		)
		if err != nil {
			return nil, nil
		}
		return []*settings.Update{
			{
				Field: &settings.Update_AllowRegister{
					AllowRegister: allowRegister == "true",
				},
			},
		}, nil
	case settingsIsVerifiedEmailRequired:
		defIsVerifiedEmailRequired := strconv.FormatBool(s.GetIsVerifiedEmailRequired())
		isVerifiedEmailRequired, err := c.Prompter.Input(
			"Verify Email Required:", common.BoolValidateOpt, common.InputDefaultOpt(defIsVerifiedEmailRequired),
		)
		if err != nil {
			return nil, nil
		}
		return []*settings.Update{
			{
				Field: &settings.Update_IsVerifiedEmailRequired{
					IsVerifiedEmailRequired: isVerifiedEmailRequired == "true",
				},
			},
		}, nil
	case settingsIsVerifiedPhoneRequired:
		defIsVerifiedPhoneRequired := strconv.FormatBool(s.GetIsVerifiedPhoneRequired())
		isVerifiedPhoneRequired, err := c.Prompter.Input(
			"Verify Phone Required:", common.BoolValidateOpt, common.InputDefaultOpt(defIsVerifiedPhoneRequired),
		)
		if err != nil {
			return nil, nil
		}
		return []*settings.Update{
			{
				Field: &settings.Update_IsVerifiedPhoneRequired{
					IsVerifiedPhoneRequired: isVerifiedPhoneRequired == "true",
				},
			},
		}, nil
	case settingsAccessTokenTTL:
		defAccessTokenTTL := strconv.Itoa(int(s.GetAccessTokenTtl().AsDuration().Minutes()))
		accessTokenTTL, err := c.Prompter.Input(
			"Access Token TTL (minutes):", common.ValidateIntOpt, common.InputDefaultOpt(defAccessTokenTTL),
		)
		if err != nil {
			return nil, nil
		}

		accessTokenTTLInt, _ := strconv.Atoi(accessTokenTTL)
		return []*settings.Update{
			{
				Field: &settings.Update_AccessTokenTtl{
					AccessTokenTtl: &durationpb.Duration{
						Seconds: int64(accessTokenTTLInt * 60),
					},
				},
			},
		}, err
	case settingsRefreshTokenTTL:
		defRefreshTokenTTL := strconv.Itoa(int(s.GetRefreshTokenTtl().AsDuration().Hours()))
		refreshTokenTTL, err := c.Prompter.Input(
			"Refresh Token TTL (hours):", common.ValidateIntOpt, common.InputDefaultOpt(defRefreshTokenTTL),
		)
		if err != nil {
			return nil, nil
		}

		refreshTokenTTLInt, _ := strconv.Atoi(refreshTokenTTL)
		return []*settings.Update{
			{
				Field: &settings.Update_RefreshTokenTtl{
					RefreshTokenTtl: &durationpb.Duration{
						Seconds: int64(refreshTokenTTLInt * 60 * 60),
					},
				},
			},
		}, nil
	case settingsVerificationCodeTTL:
		defVerificationCodeTTL := strconv.Itoa(int(s.GetVerificationCodeTtl().AsDuration().Minutes()))
		verificationCodeTTL, err := c.Prompter.Input(
			"Verification Code TTL (minutes):", common.ValidateIntOpt, common.InputDefaultOpt(defVerificationCodeTTL),
		)
		if err != nil {
			return nil, nil
		}
		verificationCodeTTLInt, _ := strconv.Atoi(verificationCodeTTL)
		return []*settings.Update{
			{
				Field: &settings.Update_VerificationCodeTtl{
					VerificationCodeTtl: &durationpb.Duration{
						Seconds: int64(verificationCodeTTLInt * 60),
					},
				},
			},
		}, nil
	case settingsPasswordHashing:
		u, err := c.getPasswordHashingUpdate(s.GetPasswordHashing())
		if err != nil {
			return nil, nil
		}
		return []*settings.Update{u}, err
	case settingsLoginMechanisms:
		u, err := c.getLoginMechanismsUpdate(s.GetLoginMechanisms())
		if err != nil {
			return nil, nil
		}
		return []*settings.Update{u}, nil
	case settingsEmailProvider:
		u, err := c.promptEmailProvider(s)
		if err != nil {
			return nil, nil
		}
		return []*settings.Update{u}, nil
	case templateEmailWelcome:
		u, err := c.promptTemplate(s.GetTemplates().GetWelcomeEmail())
		if err != nil {
			return nil, nil
		}

		return []*settings.Update{u}, nil
	case templateResetPasswordEmail:
		u, err := c.promptTemplate(s.GetTemplates().GetResetPasswordEmail())
		if err != nil {
			return nil, nil
		}

		return []*settings.Update{u}, nil
	case templateVerifyEmail:
		u, err := c.promptTemplate(s.GetTemplates().GetVerifyEmail())
		if err != nil {
			return nil, nil
		}

		return []*settings.Update{u}, nil
	default:
		return nil, nil
	}
}

func (c *Cmd) getPasswordHashingUpdate(psh *model.HashingConfig) (*settings.Update, error) {
	fmt.Println("Currend password Hashing: " + psh.String())
	fields := []string{
		"Bcrypt",
		"Scrypt",
	}
	_, res, err := c.Prompter.Select("Choose hashing algorithm", fields)
	if err != nil {
		return nil, err
	}
	if res == "Bcrypt" {
		var defCost string
		if psh.GetBcrypt() != nil {
			defCost = strconv.Itoa(int(psh.GetBcrypt().GetCost()))
		}

		bcrypt := &model.BcryptHashingConfig{}
		cost, err := c.Prompter.Input("Cost:", common.ValidateIntOpt, common.InputDefaultOpt(defCost))
		if err != nil {
			return nil, err
		}
		costInt, _ := strconv.Atoi(cost)
		bcrypt.Cost = int32(costInt)
		return &settings.Update{
			Field: &settings.Update_PasswordHashing{
				PasswordHashing: &model.HashingConfig{
					Method: &model.HashingConfig_Bcrypt{
						Bcrypt: bcrypt,
					},
				},
			},
		}, nil
	} else if res == "Scrypt" {
		defKey := ""
		defSaltSeparator := ""
		defRounds := ""
		defMemCost := ""
		defParallelism := ""
		defKeyLength := ""
		if psh.GetScrypt() != nil {
			defKey = psh.GetScrypt().GetSignerKey()
			defSaltSeparator = psh.GetScrypt().GetSaltSeparator()
			defRounds = strconv.Itoa(int(psh.GetScrypt().GetRounds()))
			defMemCost = strconv.Itoa(int(psh.GetScrypt().GetMemCost()))
			defParallelism = strconv.Itoa(int(psh.GetScrypt().GetP()))
			defKeyLength = strconv.Itoa(int(psh.GetScrypt().GetKeyLen()))
		}

		scrypt := &model.ScryptHashingConfig{}
		key, err := c.Prompter.Input("Key:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(defKey))
		if err != nil {
			return nil, err
		}
		scrypt.SignerKey = key

		saltSeparator, err := c.Prompter.Input(
			"Salt Separator:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(defSaltSeparator),
		)
		if err != nil {
			return nil, err
		}
		scrypt.SaltSeparator = saltSeparator

		rounds, err := c.Prompter.Input(
			"Rounds:", common.ValidateIntOpt, common.InputDefaultOpt(defRounds),
		)
		if err != nil {
			return nil, err
		}
		roundsInt, _ := strconv.Atoi(rounds)
		scrypt.Rounds = int32(roundsInt)

		memCost, err := c.Prompter.Input(
			"Memory Cost:", common.ValidateIntOpt, common.InputDefaultOpt(defMemCost),
		)
		if err != nil {
			return nil, err
		}
		memCostInt, _ := strconv.Atoi(memCost)
		scrypt.MemCost = int32(memCostInt)

		parallelism, err := c.Prompter.Input(
			"Parallelism:", common.ValidateIntOpt, common.InputDefaultOpt(defParallelism),
		)
		if err != nil {
			return nil, err
		}
		parallelismInt, _ := strconv.Atoi(parallelism)
		scrypt.P = int32(parallelismInt)

		keyLength, err := c.Prompter.Input(
			"Key Length:", common.ValidateIntOpt, common.InputDefaultOpt(defKeyLength),
		)
		if err != nil {
			return nil, err
		}
		keyLengthInt, _ := strconv.Atoi(keyLength)
		scrypt.KeyLen = int32(keyLengthInt)

		return &settings.Update{
			Field: &settings.Update_PasswordHashing{
				PasswordHashing: &model.HashingConfig{
					Method: &model.HashingConfig_Scrypt{
						Scrypt: scrypt,
					},
				},
			},
		}, nil
	}

	return nil, nil
}

func (c *Cmd) getLoginMechanismsUpdate(current []model.LoginType) (*settings.Update, error) {
	currentString := []string{}
	for _, l := range current {
		currentString = append(currentString, loginTypeToString(l))
	}
	fmt.Println("Current login mechanisms: ", currentString)
	fields := []string{
		loginTypeToString(model.LoginType_LOGIN_TYPE_EMAIL_PASSWORD),
		loginTypeToString(model.LoginType_LOGIN_TYPE_PHONE_PASSWORD),
		loginTypeToString(model.LoginType_LOGIN_TYPE_USERNAME_PASSWORD),
		"Done",
	}

	selected := []model.LoginType{}
	for {
		i, res, err := c.Prompter.Select("Choose login types", fields)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		if res == "Done" {
			break
		}
		loginType := loginTypeFromString(res)
		if loginType == model.LoginType_LOGIN_TYPE_UNSPECIFIED {
			fmt.Println(errors.New("invalid login type"))
			continue
		}
		selected = append(selected, loginType)
		fields = slices.Delete(fields, i, i+1)
	}

	u := &settings.Update_LoginMechanisms{
		LoginMechanisms: selected,
	}

	return &settings.Update{
		Field: &settings.Update_LoginMechanisms_{
			LoginMechanisms: u,
		},
	}, nil
}

func parseSettingsUpdate() (*settings.Update, error) {
	switch field {
	case common.FormatField(settingsAllowRegister.String()):
		allowRegister, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_AllowRegister{
				AllowRegister: allowRegister,
			},
		}, nil
	case common.FormatField(settingsIsVerifiedEmailRequired.String()):
		isVerifiedEmailRequired, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_IsVerifiedEmailRequired{
				IsVerifiedEmailRequired: isVerifiedEmailRequired,
			},
		}, nil
	case common.FormatField(settingsIsVerifiedPhoneRequired.String()):
		isVerifiedPhoneRequired, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_IsVerifiedPhoneRequired{
				IsVerifiedPhoneRequired: isVerifiedPhoneRequired,
			},
		}, nil
	case common.FormatField(settingsAccessTokenTTL.String()):
		accessTokenTTL, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_AccessTokenTtl{
				AccessTokenTtl: &durationpb.Duration{
					Seconds: int64(accessTokenTTL * 60),
				},
			},
		}, nil
	case common.FormatField(settingsRefreshTokenTTL.String()):
		refreshTokenTTL, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_RefreshTokenTtl{
				RefreshTokenTtl: &durationpb.Duration{
					Seconds: int64(refreshTokenTTL * 60 * 60),
				},
			},
		}, nil
	case common.FormatField(settingsVerificationCodeTTL.String()):
		verificationCodeTTL, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_VerificationCodeTtl{
				VerificationCodeTtl: &durationpb.Duration{
					Seconds: int64(verificationCodeTTL * 60),
				},
			},
		}, nil
	case common.FormatField(settingsPasswordHashing.String()):
		jsonValue := []byte(value)
		hashingConfig := model.HashingConfig{}
		err := protojson.Unmarshal(jsonValue, &hashingConfig)
		if err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_PasswordHashing{
				PasswordHashing: &hashingConfig,
			},
		}, nil
	case common.FormatField(settingsLoginMechanisms.String()):
		jsonValue := []byte(value)
		loginMechanisms := []model.LoginType{}
		err := json.Unmarshal(jsonValue, &loginMechanisms)
		if err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_LoginMechanisms_{
				LoginMechanisms: &settings.Update_LoginMechanisms{
					LoginMechanisms: loginMechanisms,
				},
			},
		}, nil
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

func (c *Cmd) promptEmailProvider(s *settings.Settings) (*settings.Update, error) {
	_, field, err := c.Prompter.Select("Choose a type:", []string{
		"MailJet",
		"Smtp",
		"Default",
	})
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
		if err := c.promptEmailProviderFields(prov, field); err != nil {
			return nil, err
		}
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
		if err := c.promptEmailProviderFields(prov, field); err != nil {
			return nil, err
		}
		return &settings.Update{
			Field: &settings.Update_EmailProvider{
				EmailProvider: prov,
			},
		}, nil
	default:
		return nil, nil
	}
}

func (c *Cmd) promptEmailProviderFields(p *settings.EmailProvider, prov string) error {
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
		_, res, err := c.Prompter.Select("Choose a field to update:", fields)
		if err != nil {
			return err
		}
		if res == "Done" {
			break
		}

		switch res {
		case emailProviderPublicKey.String():
			key, err := c.Prompter.Input(
				"Enter public key:", common.ValidateNonEmptyOpt,
			)
			if err != nil {
				return err
			}
			p.Credentials.PublicKey = key
		case emailProviderPrivateKey.String():
			key, err := c.Prompter.Input(
				"Enter private key:", common.ValidateNonEmptyOpt,
			)
			if err != nil {
				return err
			}
			p.Credentials.PrivateKey = key
		case emailProviderFromEmail.String():
			email, err := c.Prompter.Input(
				"Enter from email:",
				common.ValidateEmailOpt,
				common.InputDefaultOpt(p.GetFrom()),
			)
			if err != nil {
				return err
			}
			p.From = email
		case emailProviderHost.String():
			host, err := c.Prompter.Input(
				"Enter host:",
				common.ValidateNonEmptyOpt,
				common.InputDefaultOpt(p.GetInstance().GetSmtp().GetHost()),
			)
			if err != nil {
				return err
			}
			p.GetInstance().GetSmtp().Host = host
		case emailProviderPort.String():
			port, err := c.Prompter.Input(
				"Enter port:",
				common.ValidateNonEmptyOpt,
				common.InputDefaultOpt(strconv.Itoa(int(p.GetInstance().GetSmtp().GetPort()))),
			)
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

func (c *Cmd) promptTemplate(t *settings.Template) (*settings.Update, error) {
	fields := []string{
		tempalteFieldSubject.String(),
		templateFieldBody.String(),
		"Done",
	}

	for {
		_, res, err := c.Prompter.Select("Choose a field to update:", fields)
		if err != nil {
			return nil, err
		}
		if res == "Done" {
			break
		}

		switch res {
		case tempalteFieldSubject.String():
			subject, err := c.Prompter.Input(
				"Enter subject:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(t.GetSubject()),
			)
			if err != nil {
				return nil, err
			}
			t.Subject = subject
		case templateFieldBody.String():
			body, err := c.Prompter.Input(
				"Enter body:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(t.GetBody()),
			)
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
