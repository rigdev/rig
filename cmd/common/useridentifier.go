package common

import (
	"fmt"
	"net/mail"

	"github.com/nyaruka/phonenumbers"
	"github.com/rigdev/rig-go-api/model"
)

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

func UserIndentifier(p Prompter) (*model.UserIdentifier, error) {
	var err error
	identifierStr, err := p.Input("Username or email:", ValidateAllOpt)
	if err != nil {
		return nil, err
	}
	identifier, err := ParseUserIdentifier(identifierStr)
	if err != nil {
		return nil, err
	}
	return identifier, nil
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
