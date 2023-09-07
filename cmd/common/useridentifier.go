package common

import (
	"fmt"
	"net/mail"

	"github.com/nyaruka/phonenumbers"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
)

func PromptUserIndentifierUpdate() (*user.Update, error) {
	var err error
	identifier, err := PromptGetInput("Username, email or phone number:", ValidateAll)
	if err != nil {
		return nil, err
	}
	update, err := ParseUserIdentifierUpdate(identifier)
	if err != nil {
		return nil, err
	}
	return update, nil
}

func PromptUserIndentifier() (*model.UserIdentifier, error) {
	var err error
	identifierStr, err := PromptGetInput("Username, email or phone number:", ValidateAll)
	if err != nil {
		return nil, err
	}
	identifier, err := ParseUserIdentifier(identifierStr)
	if err != nil {
		return nil, err
	}
	return identifier, nil
}

func GetUserIdentifierUpdates(username, email, phoneNumber string) ([]*user.Update, error) {
	if username == "" && email == "" && phoneNumber == "" {
		update, err := PromptUserIndentifierUpdate()
		if err != nil {
			return nil, err
		}
		return []*user.Update{update}, nil
	}

	var updates []*user.Update
	if username != "" {
		updates = append(updates, makeUpdateUsername(username))
	}
	if email != "" {
		updates = append(updates, makeUpdateUsername(email))
	}
	if phoneNumber != "" {
		updates = append(updates, makeUpdateUsername(phoneNumber))
	}

	return updates, nil
}

func GetUserIdentifier(username, email, phoneNumber string) (*model.UserIdentifier, error) {
	if username == "" && email == "" && phoneNumber == "" {
		identifier, err := PromptUserIndentifier()
		if err != nil {
			return nil, err
		}
		return identifier, nil
	}

	if username != "" {
		return makeIdentifierUsername(username), nil
	}
	if email != "" {
		return makeIdentifierEmail(email), nil
	}
	if phoneNumber != "" {
		return makeIdentifierPhoneNumber(phoneNumber), nil
	}

	// We should not get here
	return nil, fmt.Errorf("something went wrong")
}

func GetUserAndPasswordUpdates(username, email, phoneNumber, password string) ([]*user.Update, error) {
	updates, err := GetUserIdentifierUpdates(username, email, phoneNumber)
	if err != nil {
		return nil, err
	}

	if password == "" {
		password, err = GetPasswordPrompt("Password:")
		if err != nil {
			return nil, err
		}
	}

	updates = append(updates, makeUpdatePassword(password))

	return updates, nil
}

func ParseUserIdentifier(identifier string) (*model.UserIdentifier, error) {
	var err error
	if _, err = mail.ParseAddress(identifier); err == nil {
		return makeIdentifierEmail(identifier), nil
	} else if _, err = phonenumbers.Parse(identifier, ""); err == nil {
		return makeIdentifierPhoneNumber(identifier), nil
	} else if err = ValidateSystemName(identifier); err == nil {
		return makeIdentifierUsername(identifier), nil
	}
	return nil, fmt.Errorf("invalid user identifier")
}

func ParseUserIdentifierUpdate(identifier string) (*user.Update, error) {
	var err error
	if _, err = mail.ParseAddress(identifier); err == nil {
		return makeUpdateEmail(identifier), nil
	} else if _, err = phonenumbers.Parse(identifier, ""); err == nil {
		return makeUpdatePhoneNumber(identifier), nil
	} else if err = ValidateSystemName(identifier); err == nil {
		return makeUpdateUsername(identifier), nil
	}
	return nil, fmt.Errorf("invalid user identifier")
}

func makeUpdateUsername(username string) *user.Update {
	return &user.Update{Field: &user.Update_Username{Username: username}}
}

func makeUpdateEmail(email string) *user.Update {
	return &user.Update{Field: &user.Update_Email{Email: email}}
}

func makeUpdatePhoneNumber(phoneNumber string) *user.Update {
	return &user.Update{Field: &user.Update_PhoneNumber{PhoneNumber: phoneNumber}}
}

func makeUpdatePassword(password string) *user.Update {
	return &user.Update{Field: &user.Update_Password{Password: password}}
}

func makeIdentifierUsername(username string) *model.UserIdentifier {
	return &model.UserIdentifier{Identifier: &model.UserIdentifier_Username{Username: username}}
}

func makeIdentifierEmail(email string) *model.UserIdentifier {
	return &model.UserIdentifier{Identifier: &model.UserIdentifier_Email{Email: email}}
}

func makeIdentifierPhoneNumber(phoneNumber string) *model.UserIdentifier {
	return &model.UserIdentifier{Identifier: &model.UserIdentifier_PhoneNumber{PhoneNumber: phoneNumber}}
}
