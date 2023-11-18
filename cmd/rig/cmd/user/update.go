package user

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type (
	userField        int64
	userProfileField int64
)

const (
	userUndefined userField = iota
	userEmail
	userUsername
	userPhoneNumber
	userPassword
	userProfile
	userIsEmailVerified
	userIsPhoneVerified
	userResetSessions
	userSetMetaData
	userDeleteMetaData
)

const (
	userProfileUndefined userProfileField = iota
	userProfileFirstName
	userProfileLastName
	userProfilePreferredLanguage
	userProfileCountry
)

func (f userField) String() string {
	switch f {
	case userEmail:
		return "Email"
	case userUsername:
		return "Username"
	case userPhoneNumber:
		return "Phone number"
	case userPassword:
		return "Password"
	case userProfile:
		return "Profile"
	case userIsEmailVerified:
		return "Email verified"
	case userIsPhoneVerified:
		return "Phone verified"
	case userResetSessions:
		return "Reset sessions"
	case userSetMetaData:
		return "Set metadata"
	case userDeleteMetaData:
		return "Delete metadata"
	default:
		return "Unknown"
	}
}

func (f userProfileField) String() string {
	switch f {
	case userProfileFirstName:
		return "First name"
	case userProfileLastName:
		return "Last name"
	case userProfilePreferredLanguage:
		return "Preferred language"
	case userProfileCountry:
		return "Country"
	default:
		return "Unknown"
	}
}

func (c *Cmd) update(ctx context.Context, cmd *cobra.Command, args []string) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	u, id, err := common.GetUser(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	if value != "" && field != "" {
		u, err := parseUpdate()
		if err != nil {
			return err
		}

		_, err = c.Rig.User().Update(ctx, &connect.Request[user.UpdateRequest]{
			Msg: &user.UpdateRequest{
				UserId:  id,
				Updates: []*user.Update{u},
			},
		})
		if err != nil {
			return err
		}

		cmd.Printf("Successfully updated user %s\n", identifier)
		return nil
	}

	fields := []string{
		userEmail.String(),
		userUsername.String(),
		userPhoneNumber.String(),
		userPassword.String(),
		userProfile.String(),
		userIsEmailVerified.String(),
		userIsPhoneVerified.String(),
		userResetSessions.String(),
		userSetMetaData.String(),
		userDeleteMetaData.String(),
		"Done",
	}

	updates := []*user.Update{}
	for {
		i, res, err := common.PromptSelect("Choose a field to update:", fields)
		if err != nil {
			return err
		}
		if res == "Done" {
			break
		}
		u, err := promptUserUpdate(userField(i+1), u)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if u != nil {
			updates = append(updates, u)
		}
	}

	_, err = c.Rig.User().Update(ctx, connect.NewRequest(&user.UpdateRequest{
		UserId:  id,
		Updates: updates,
	}))
	if err != nil {
		return err
	}

	cmd.Println("Updated user")
	return nil
}

func promptUserUpdate(f userField, u *user.User) (*user.Update, error) {
	switch f {
	case userEmail:
		defEmail := u.GetUserInfo().GetEmail()
		email, err := common.PromptInput("Email:", common.ValidateEmailOpt, common.InputDefaultOpt(defEmail))
		if err != nil {
			return nil, nil
		}

		if email != defEmail {
			return &user.Update{
				Field: &user.Update_Email{
					Email: email,
				},
			}, nil
		}
	case userUsername:
		defUsername := u.GetUserInfo().GetUsername()
		username, err := common.PromptInput("Username:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(defUsername))
		if err != nil {
			return nil, nil
		}

		if username != defUsername {
			return &user.Update{
				Field: &user.Update_Username{
					Username: username,
				},
			}, nil
		}
	case userPhoneNumber:
		defPhone := u.GetUserInfo().GetPhoneNumber()
		phone, err := common.PromptInput("Phone:", common.ValidatePhoneOpt, common.InputDefaultOpt(defPhone))
		if err != nil {
			return nil, nil
		}

		if phone != defPhone {
			return &user.Update{
				Field: &user.Update_PhoneNumber{
					PhoneNumber: phone,
				},
			}, nil
		}
	case userPassword:
		password, err := common.PromptPassword("Password:")
		if err != nil {
			return nil, nil
		}
		return &user.Update{
			Field: &user.Update_Password{
				Password: password,
			},
		}, nil
	case userIsEmailVerified:
		defIsEmailVerified := strconv.FormatBool(u.GetIsEmailVerified())
		isEmailVerified, err := common.PromptInput("Is email verified:", common.BoolValidateOpt, common.InputDefaultOpt(defIsEmailVerified))
		if err != nil {
			return nil, nil
		}
		if isEmailVerified != defIsEmailVerified {
			return &user.Update{
				Field: &user.Update_IsEmailVerified{
					IsEmailVerified: isEmailVerified == "true",
				},
			}, nil
		}
	case userIsPhoneVerified:
		defIsPhoneVerified := strconv.FormatBool(u.GetIsPhoneVerified())
		isPhoneVerified, err := common.PromptInput("Is phone verified:", common.BoolValidateOpt, common.InputDefaultOpt(defIsPhoneVerified))
		if err != nil {
			return nil, nil
		}
		if isPhoneVerified != defIsPhoneVerified {
			return &user.Update{
				Field: &user.Update_IsPhoneVerified{
					IsPhoneVerified: isPhoneVerified == "true",
				},
			}, nil
		}
	case userResetSessions:
		return &user.Update{
			Field: &user.Update_ResetSessions_{},
		}, nil
	case userProfile:
		u, err := getUserProfileUpdate(u.GetProfile())
		if err != nil {
			return nil, nil
		}
		return u, err
	case userSetMetaData:
		key, err := common.PromptInput("Key:", common.ValidateNonEmptyOpt)
		if err != nil {
			return nil, nil
		}
		value, err := common.PromptInput("Value:", common.ValidateNonEmptyOpt)
		if err != nil {
			return nil, nil
		}
		return &user.Update{
			Field: &user.Update_SetMetadata{
				SetMetadata: &model.Metadata{
					Key:   key,
					Value: []byte(value),
				},
			},
		}, nil
	case userDeleteMetaData:
		key, err := common.PromptInput("Key:", common.ValidateNonEmptyOpt)
		if err != nil {
			return nil, nil
		}

		return &user.Update{
			Field: &user.Update_DeleteMetadataKey{
				DeleteMetadataKey: key,
			},
		}, nil
	default:
		return nil, nil
	}
	return nil, nil
}

func getUserProfileUpdate(p *user.Profile) (*user.Update, error) {
	fields := []string{
		userProfileFirstName.String(),
		userProfileLastName.String(),
		userProfilePreferredLanguage.String(),
		userProfileCountry.String(),
		"Done",
	}

	pp := &user.Profile{
		FirstName:         p.GetFirstName(),
		LastName:          p.GetLastName(),
		PreferredLanguage: p.GetPreferredLanguage(),
		Country:           p.GetCountry(),
	}
	for {
		i, res, err := common.PromptSelect("Choose a field to update:", fields)
		if err != nil {
			return nil, nil
		}
		if res == "Done" {
			break
		}
		err = promptUserProfileUpdate(userProfileField(i+1), pp)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
	}
	if proto.Equal(p, pp) {
		return nil, nil
	}

	return &user.Update{
		Field: &user.Update_Profile{
			Profile: pp,
		},
	}, nil
}

func promptUserProfileUpdate(f userProfileField, p *user.Profile) error {
	switch f {
	case userProfileFirstName:
		defFirstName := p.GetFirstName()
		firstName, err := common.PromptInput("First name:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(defFirstName))
		if err != nil {
			return err
		}
		if firstName != defFirstName {
			p.FirstName = firstName
		}
	case userProfileLastName:
		defLastName := p.GetLastName()
		lastName, err := common.PromptInput("Last name:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(defLastName))
		if err != nil {
			return err
		}
		if lastName != defLastName {
			p.LastName = lastName
		}
	case userProfilePreferredLanguage:
		defPreferredLanguage := p.GetPreferredLanguage()
		preferredLanguage, err := common.PromptInput("Preferred language:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(defPreferredLanguage))
		if err != nil {
			return err
		}
		if preferredLanguage != defPreferredLanguage {
			p.PreferredLanguage = preferredLanguage
		}
	case userProfileCountry:
		defCountry := p.GetCountry()
		country, err := common.PromptInput("Country:", common.ValidateNonEmptyOpt, common.InputDefaultOpt(defCountry))
		if err != nil {
			return err
		}
		if country != defCountry {
			p.Country = country
		}
	}
	return nil
}

func parseUpdate() (*user.Update, error) {
	switch field {
	case common.FormatField(userEmail.String()):
		return &user.Update{
			Field: &user.Update_Email{
				Email: value,
			},
		}, nil
	case common.FormatField(userUsername.String()):
		return &user.Update{
			Field: &user.Update_Username{
				Username: value,
			},
		}, nil
	case common.FormatField(userPhoneNumber.String()):
		return &user.Update{
			Field: &user.Update_PhoneNumber{
				PhoneNumber: value,
			},
		}, nil
	case common.FormatField(userPassword.String()):
		return &user.Update{
			Field: &user.Update_Password{
				Password: value,
			},
		}, nil
	case common.FormatField(userProfile.String()):
		jsonValue := []byte(value)
		p := user.Profile{}
		if err := protojson.Unmarshal(jsonValue, &p); err != nil {
			return nil, err
		}
		return &user.Update{
			Field: &user.Update_Profile{
				Profile: &p,
			},
		}, nil
	case common.FormatField(userIsEmailVerified.String()):
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return &user.Update{
			Field: &user.Update_IsEmailVerified{
				IsEmailVerified: b,
			},
		}, nil
	case common.FormatField(userIsPhoneVerified.String()):
		b, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
		return &user.Update{
			Field: &user.Update_IsPhoneVerified{
				IsPhoneVerified: b,
			},
		}, nil
	case common.FormatField(userResetSessions.String()):
		return &user.Update{
			Field: &user.Update_ResetSessions_{},
		}, nil
	case common.FormatField(userSetMetaData.String()):
		// temp struct to keep a key value pair
		keyValue := struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{}

		jsonValue := []byte(value)
		if err := json.Unmarshal(jsonValue, &keyValue); err != nil {
			return nil, err
		}
		return &user.Update{
			Field: &user.Update_SetMetadata{
				SetMetadata: &model.Metadata{
					Key:   keyValue.Key,
					Value: []byte(keyValue.Value),
				},
			},
		}, nil
	case common.FormatField(userDeleteMetaData.String()):
		return &user.Update{
			Field: &user.Update_DeleteMetadataKey{
				DeleteMetadataKey: value,
			},
		}, nil
	default:
		return nil, errors.InvalidArgumentErrorf("Unknown field")
	}
}
