package user

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/rigdev/rig/pkg/errors"
	utils2 "github.com/rigdev/rig/pkg/utils"
	"github.com/spf13/cobra"
)

func UserCreate(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	var err error

	if email == "" && phoneNumber == "" && username == "" {
		email, err = utils.PromptGetInput("Email:", utils2.ValidateEmail)
		if err != nil {
			return err
		}

		phoneNumber, err = utils.PromptGetInput("Phone number:", utils2.ValidatePhone)
		if err != nil {
			return err
		}

		usernameValidate := func(username string) error {
			if username == "" && phoneNumber == "" && email == "" {
				return errors.InvalidArgumentErrorf("Please provide atleast one identifier")
			}
			return nil
		}

		username, err = utils.PromptGetInput("Username:", usernameValidate)
		if err != nil {
			return err
		}
	}

	if password == "" {
		password, err = utils.GetPasswordPrompt("Password:")
		if err != nil {
			return err
		}
	} else if err := utils2.ValidatePassword(password); err != nil {
		return err
	}

	updates := []*user.Update{}
	if username != "" {
		updates = append(updates, &user.Update{
			Field: &user.Update_Username{
				Username: username,
			},
		})
	}
	if phoneNumber != "" {
		updates = append(updates, &user.Update{
			Field: &user.Update_PhoneNumber{
				PhoneNumber: phoneNumber,
			},
		})
	}
	if email != "" {
		updates = append(updates, &user.Update{
			Field: &user.Update_Email{
				Email: email,
			},
		})
	}
	updates = append(updates, &user.Update{
		Field: &user.Update_Password{
			Password: password,
		},
	})

	res, err := nc.User().Create(ctx, &connect.Request[user.CreateRequest]{
		Msg: &user.CreateRequest{
			Initializers: updates,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("User created with ID:", res.Msg.GetUser().GetUserId())
	return nil
}
