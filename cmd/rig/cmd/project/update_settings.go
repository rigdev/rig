package project

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	settingsField int32
)

const (
	settingsUndefined settingsField = iota
	settingsAddDockerRegistry
	settingsDeleteDockerRegistry
)

func (f settingsField) String() string {
	switch f {
	case settingsAddDockerRegistry:
		return "Add Docker Registry"
	case settingsDeleteDockerRegistry:
		return "Delete Docker Registry"
	default:
		return "Undefined"
	}
}

func (c *Cmd) updateSettings(ctx context.Context, cmd *cobra.Command, _ []string) error {
	res, err := c.Rig.ProjectSettings().GetSettings(ctx, &connect.Request[settings.GetSettingsRequest]{})
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

		_, err = c.Rig.ProjectSettings().UpdateSettings(ctx, &connect.Request[settings.UpdateSettingsRequest]{
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
		settingsAddDockerRegistry.String(),
		settingsDeleteDockerRegistry.String(),
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
			updates = append(updates, u)
		}
	}

	if len(updates) == 0 {
		cmd.Println("No settings updated")
		return nil
	}

	_, err = c.Rig.ProjectSettings().UpdateSettings(ctx, &connect.Request[settings.UpdateSettingsRequest]{
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

func (c *Cmd) promptSettingsUpdate(f settingsField, s *settings.Settings) (*settings.Update, error) {
	switch f {
	case settingsAddDockerRegistry:
		return c.promptAddDockerRegistry()
	case settingsDeleteDockerRegistry:
		return c.promptDeleteDockerRegistry(s)
	default:
		return nil, nil
	}
}

func (c *Cmd) promptDeleteDockerRegistry(s *settings.Settings) (*settings.Update, error) {
	if len(s.GetDockerRegistries()) == 0 {
		return nil, nil
	}

	var hosts []string
	for _, r := range s.GetDockerRegistries() {
		hosts = append(hosts, r.GetHost())
	}

	_, res, err := c.Prompter.Select("Choose a registry to delete:", hosts)
	if err != nil {
		return nil, err
	}

	return &settings.Update{
		Field: &settings.Update_DeleteDockerRegistry{
			DeleteDockerRegistry: res,
		},
	}, nil
}

func (c *Cmd) promptAddDockerRegistry() (*settings.Update, error) {
	host, err := c.Prompter.Input("Enter host:", common.ValidateNonEmptyOpt)
	if err != nil {
		return nil, err
	}

	username, err := c.Prompter.Input("Enter username:", common.ValidateNonEmptyOpt)
	if err != nil {
		return nil, err
	}

	password, err := c.Prompter.Input("Enter password:", common.ValidateNonEmptyOpt)
	if err != nil {
		return nil, err
	}

	email, err := c.Prompter.Input("Enter email:", common.ValidateEmailOpt)
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

func parseSettingsUpdate() (*settings.Update, error) {
	switch field {
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
	default:
		return nil, nil
	}
}
