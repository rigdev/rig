package user

import (
	"context"
	"fmt"
	"net/mail"

	"connectrpc.com/connect"
	"github.com/nyaruka/phonenumbers"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, _ []string) error {
	updates, err := c.getUserAndPasswordUpdates(username, email, password)
	if err != nil {
		return err
	}

	if role == "" {
		_, role, err = c.Prompter.Select("What is the role of the user?",
			[]string{"admin", "owner", "developer", "viewer"})
		if err != nil {
			return err
		}
	}

	res, err := c.Rig.User().Create(ctx, &connect.Request[user.CreateRequest]{
		Msg: &user.CreateRequest{
			Initializers:   updates,
			InitialGroupId: role,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Succesfully created user with ID:", res.Msg.GetUser().GetUserId())

	return nil
}

func (c *Cmd) getUserAndPasswordUpdates(username, email, password string) ([]*user.Update, error) {
	updates, err := c.getUserIdentifierUpdates(username, email)
	if err != nil {
		return nil, err
	}

	if password == "" {
		password, err = c.Prompter.Password("Password:")
		if err != nil {
			return nil, err
		}
	}

	updates = append(updates, makeUpdatePassword(password))

	return updates, nil
}

func (c *Cmd) getUserIdentifierUpdates(username, email string) ([]*user.Update, error) {
	if username == "" && email == "" {
		update, err := c.userIndentifierUpdate()
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

	return updates, nil
}

func (c *Cmd) userIndentifierUpdate() (*user.Update, error) {
	var err error
	identifier, err := c.Prompter.Input("Username or email:", common.ValidateAllOpt)
	if err != nil {
		return nil, err
	}
	update, err := parseUserIdentifierUpdate(identifier)
	if err != nil {
		return nil, err
	}
	return update, nil
}

func (c *Cmd) getUserIdentifier(username, email, phoneNumber string) (*model.UserIdentifier, error) {
	if username == "" && email == "" && phoneNumber == "" {
		identifier, err := common.UserIndentifier(c.Prompter)
		if err != nil {
			return nil, err
		}
		return identifier, nil
	}

	if username != "" {
		return &model.UserIdentifier{Identifier: &model.UserIdentifier_Username{Username: username}}, nil
	}
	if email != "" {
		return &model.UserIdentifier{Identifier: &model.UserIdentifier_Email{Email: email}}, nil
	}
	if phoneNumber != "" {
		return &model.UserIdentifier{Identifier: &model.UserIdentifier_PhoneNumber{PhoneNumber: phoneNumber}}, nil
	}

	// We should not get here
	return nil, fmt.Errorf("something went wrong")
}

func parseUserIdentifierUpdate(identifier string) (*user.Update, error) {
	var err error
	if _, err = mail.ParseAddress(identifier); err == nil {
		return makeUpdateEmail(identifier), nil
	} else if _, err = phonenumbers.Parse(identifier, ""); err == nil {
		return makeUpdatePhoneNumber(identifier), nil
	} else if err = common.ValidateSystemName(identifier); err == nil {
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
