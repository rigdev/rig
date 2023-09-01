package utils

import (
	"context"
	goerrors "errors"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/docker/distribution/reference"
	"github.com/manifoldco/promptui"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/nyaruka/phonenumbers"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var GetInputTemplates = &promptui.PromptTemplates{
	Prompt:  "{{ . }} ",
	Valid:   "{{ . | green }} ",
	Invalid: "{{ . | red }} ",
	Success: "{{ . | bold }} ",
}

var ValidateAll = func(input string) error {
	return nil
}

var BoolValidate = func(bool string) error {
	if bool != "true" && bool != "false" {
		return errors.InvalidArgumentErrorf("invalid boolean value")
	}
	return nil
}

var ValidateInt = func(input string) error {
	_, err := strconv.Atoi(input)
	if err != nil {
		return err
	}
	return nil
}

var ValidateNonEmpty = func(input string) error {
	if input == "" {
		return errors.InvalidArgumentErrorf("value cannot be empty")
	}
	return nil
}

var ValidateEmail = func(input string) error {
	_, err := mail.ParseAddress(input)
	if err != nil {
		return err
	}
	return nil
}

var ValidateSystemName = func(input string) error {
	if l := len(input); l < 3 || l > 63 {
		return errors.InvalidArgumentErrorf("must be between 3 and 63 characters long")
	}

	if !regexp.MustCompile(`^[a-z][a-z0-9-]+[a-z0-9]$`).MatchString(input) {
		return errors.InvalidArgumentErrorf("invalid name; can only contain a-z, 0-9 and '-'")
	}

	return nil
}

var ValidateURL = func(input string) error {
	_, err := url.Parse(input)
	return err
}

var ValidateImage = func(input string) error {
	_, err := reference.ParseDockerRef(input)
	if err != nil {
		return err
	}

	return nil
}

func PromptGetInput(label string, validate func(input string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:     label,
		Templates: GetInputTemplates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

func parseBool(s string) (bool, error) {
	switch s {
	case "1", "t", "T", "true", "TRUE", "True", "y", "Y", "yes", "YES", "Yes":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "n", "N", "no", "NO", "No":
		return false, nil
	}
	return false, errors.InvalidArgumentErrorf("invalid bool format")
}

func PromptConfirm(label string, def bool) (bool, error) {
	prompt := promptui.Prompt{
		Label: label,
		// Templates: GetInputTemplates,
		IsConfirm: true,
		Validate: func(s string) error {
			if s == "" {
				return nil
			}

			if _, err := parseBool(s); err != nil {
				return err
			}

			return nil
		},
		Default: "N",
	}

	if def {
		prompt.Default = "Y"
	}

	result, err := prompt.Run()
	confirmed := !goerrors.Is(err, promptui.ErrAbort)
	if err != nil && confirmed {
		return false, err
	}

	if result == "" {
		return def, nil
	}

	return parseBool(result)
}

func PromptGetInputWithDefault(label string, validate func(input string) error, def string) (string, error) {
	prompt := promptui.Prompt{
		Label:     label,
		Templates: GetInputTemplates,
		Validate:  validate,
		Default:   def,
		AllowEdit: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

func PromptSelect(label string, items []string, hideSelected bool) (int, string, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "->{{ . | cyan }}",
		Inactive: "  {{ . | cyan }}",
		Selected: "{{ . | green }}",
	}

	prompt := promptui.Select{
		Templates:    templates,
		Label:        label,
		Items:        items,
		HideSelected: hideSelected,
	}

	i, res, err := prompt.Run()
	if err != nil {
		return 0, "", err
	}

	return i, res, nil
}

func GetPasswordPrompt(label string) (string, error) {
	prompt := promptui.Prompt{
		Label:       label,
		Templates:   GetInputTemplates,
		HideEntered: true,
		Mask:        '*',
		Validate: func(input string) error {
			if err := utils.ValidatePassword(input); err != nil {
				return err
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func GetUser(ctx context.Context, identifier string, nc rig.Client) (*user.User, string, error) {
	var err error
	if identifier == "" {
		validateIdentifier := func(identifier string) error {
			if identifier == "" {
				return errors.InvalidArgumentErrorf("Please provide an identifier")
			}
			return nil
		}
		identifier, err = PromptGetInput("User Identifier:", validateIdentifier)
		if err != nil {
			return nil, "", err
		}
	}
	var u *user.User
	var resId string
	id, err := uuid.Parse(identifier)
	if err != nil {
		ident := parseUserIdentifier(identifier)
		res, err := nc.User().GetByIdentifier(ctx, connect.NewRequest(&user.GetByIdentifierRequest{
			Identifier: ident,
		}))
		if err != nil {
			return nil, "", err
		}
		resId = res.Msg.GetUser().GetUserId()
		u = res.Msg.GetUser()
	} else {
		res, err := nc.User().Get(ctx, connect.NewRequest(&user.GetRequest{
			UserId: id.String(),
		}))
		if err != nil {
			return nil, "", err
		}

		u = res.Msg.GetUser()
		resId = id.String()
	}
	return u, resId, nil
}

func parseUserIdentifier(identifier string) *model.UserIdentifier {
	var id *model.UserIdentifier
	if _, err := mail.ParseAddress(identifier); err == nil {
		id = &model.UserIdentifier{
			Identifier: &model.UserIdentifier_Email{
				Email: identifier,
			},
		}
	} else if _, err := phonenumbers.Parse(identifier, ""); err == nil {
		id = &model.UserIdentifier{
			Identifier: &model.UserIdentifier_PhoneNumber{
				PhoneNumber: identifier,
			},
		}
	} else {
		id = &model.UserIdentifier{
			Identifier: &model.UserIdentifier_Username{
				Username: identifier,
			},
		}
	}
	return id
}

func GetGroup(ctx context.Context, identifier string, nc rig.Client) (*group.Group, string, error) {
	var err error
	if identifier == "" {
		validateIdentifier := func(identifier string) error {
			if identifier == "" {
				return errors.InvalidArgumentErrorf("Please provide an identifier")
			}
			return nil
		}
		identifier, err = PromptGetInput("Group Identifier:", validateIdentifier)
		if err != nil {
			return nil, "", err
		}
	}
	var g *group.Group
	var resId string
	id, err := uuid.Parse(identifier)
	if err != nil {
		res, err := nc.Group().GetByName(ctx, connect.NewRequest(&group.GetByNameRequest{
			Name: identifier,
		}))
		if err != nil {
			return nil, "", err
		}
		resId = res.Msg.GetGroup().GetGroupId()
		g = res.Msg.GetGroup()
	} else {
		res, err := nc.Group().Get(ctx, connect.NewRequest(&group.GetRequest{
			GroupId: id.String(),
		}))
		if err != nil {
			return nil, "", err
		}
		resId = id.String()
		g = res.Msg.GetGroup()
	}
	return g, resId, nil
}

func GetDatabase(ctx context.Context, identifier string, nc rig.Client) (*database.Database, string, error) {
	var err error
	if identifier == "" {
		validateIdentifier := func(identifier string) error {
			if identifier == "" {
				return errors.InvalidArgumentErrorf("Please provide an identifier")
			}
			return nil
		}
		identifier, err = PromptGetInput("DB Identifier:", validateIdentifier)
		if err != nil {
			return nil, "", err
		}
	}
	var d *database.Database
	var id uuid.UUID
	id, err = uuid.Parse(identifier)
	var resId string
	if err != nil {
		res, err := nc.Database().GetByName(ctx, connect.NewRequest(&database.GetByNameRequest{
			Name: identifier,
		}))
		if err != nil {
			return nil, "", err
		}
		resId = res.Msg.GetDatabase().GetDatabaseId()
		d = res.Msg.GetDatabase()
	} else {
		res, err := nc.Database().Get(ctx, connect.NewRequest(&database.GetRequest{
			DatabaseId: id.String(),
		}))
		if err != nil {
			return nil, "", err
		}
		resId = id.String()
		d = res.Msg.GetDatabase()
	}
	return d, resId, nil
}

func GetStorageProvider(ctx context.Context, identifier string, nc rig.Client) (*storage.Provider, string, error) {
	var err error
	if identifier == "" {
		validateIdentifier := func(identifier string) error {
			if identifier == "" {
				return errors.InvalidArgumentErrorf("Please provide an identifier")
			}
			return nil
		}
		identifier, err = PromptGetInput("Provider Identifier:", validateIdentifier)
		if err != nil {
			return nil, "", err
		}
	}
	var p *storage.Provider
	var resId string
	id, err := uuid.Parse(identifier)
	if err != nil {
		res, err := nc.Storage().LookupProvider(ctx, connect.NewRequest(&storage.LookupProviderRequest{
			Name: identifier,
		}))
		if err != nil {
			return nil, "", err
		}
		resId = res.Msg.GetProviderId()
		p = res.Msg.GetProvider()
	} else {
		res, err := nc.Storage().GetProvider(ctx, connect.NewRequest(&storage.GetProviderRequest{
			ProviderId: id.String(),
		}))
		if err != nil {
			return nil, "", err
		}
		resId = id.String()

		p = res.Msg.GetProvider()
	}
	return p, resId, nil
}

func FormatField(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}

func ProtoToPrettyJson(m protoreflect.ProtoMessage) string {
	return protojson.Format(m)
}
