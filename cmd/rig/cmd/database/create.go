package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func Create(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	var err error
	if name == "" {
		name, err = utils.PromptGetInput("Database name:", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	if dbTypeString == "" {
		_, dbTypeString, err = utils.PromptSelect("Database type:", []string{"mongo", "postgres"}, false)
		if err != nil {
			return err
		}
	}

	if clientID == "" {
		clientID, err = utils.PromptGetInput("Client ID:", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	if clientSecret == "" {
		clientSecret, err = utils.PromptGetInput("Client secret:", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	if host == "" {
		host, err = utils.PromptGetInput("Host:", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	var config database.Config
	switch dbTypeString {
	case "mongo":
		config = database.Config{
			Config: &database.Config_Mongo{
				Mongo: &database.MongoConfig{
					Credentials: &model.ProviderCredentials{
						PublicKey:  clientID,
						PrivateKey: clientSecret,
					},
					Host: host,
				},
			},
		}
	case "postgres":
		config = database.Config{
			Config: &database.Config_Postgres{
				Postgres: &database.PostgresConfig{
					Credentials: &model.ProviderCredentials{
						PublicKey:  clientID,
						PrivateKey: clientSecret,
					},
					Host: host,
				},
			},
		}
	}

	res, err := nc.Database().Create(ctx, &connect.Request[database.CreateRequest]{Msg: &database.CreateRequest{
		Name:       name,
		Config:     &config,
		LinkTables: linkTables,
	}})
	if err != nil {
		return err
	}

	cmd.Printf("created database %s of type %s with id %s\n", name, dbTypeString, res.Msg.GetDatabase().GetDatabaseId())
	return nil
}
