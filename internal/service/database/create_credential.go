package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) CreateCredential(ctx context.Context, credentialName string, databaseID uuid.UUID) (clientId string, clientSecret string, err error) {
	db, err := s.Get(ctx, databaseID)
	if err != nil {
		return "", "", err
	}
	if credentialName == "" {
		return "", "", errors.InvalidArgumentErrorf("credential name cannot be empty")
	}

	if db.GetInfo().GetCredentials() == nil {
		db.Info.Credentials = []*database.Credential{}
	}

	for _, credential := range db.GetInfo().GetCredentials() {
		if credential.GetName() == credentialName {
			return "", "", errors.AlreadyExistsErrorf("credential with name %s already exists", credentialName)
		}
	}

	certificateID := uuid.New()

	clientSecretRaw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, clientSecretRaw); err != nil {
		return "", "", err
	}

	clientSecret = fmt.Sprint("secret_", hex.EncodeToString(clientSecretRaw))
	clientID := formatClientID(certificateID)

	switch db.GetType() {
	case database.Type_TYPE_MONGO:
		if err := s.mongoEnabled(); err != nil {
			return "", "", err
		}
		if err := s.mongo.Database("admin").RunCommand(context.Background(), bson.D{
			{Key: "createUser", Value: clientID},
			{Key: "pwd", Value: clientSecret},
			{Key: "roles", Value: []bson.M{{"role": "readWrite", "db": formatDatabaseID(databaseID)}}},
		}); err.Err() != nil {
			return "", "", err.Err()
		}
	case database.Type_TYPE_POSTGRES:
		if err := s.postgresEnabled(); err != nil {
			return "", "", err
		}
		// TODO(Oscar): (First step might successed, but second step fail, before trying again)
		if _, err := s.postgres.Exec(fmt.Sprintf("create user %s with encrypted password '%s'", clientID, clientSecret)); err != nil {
			return "", "", err
		}
		if _, err := s.postgres.Exec(fmt.Sprintf("grant all privileges on database %s to %s", formatDatabaseID(databaseID), clientID)); err != nil {
			return "", "", err
		}
	default:
		return "", "", errors.InvalidArgumentErrorf("invalid database type: %v", db.GetType())
	}
	c := &database.Credential{
		Name:      credentialName,
		ClientId:  clientID,
		CreatedAt: timestamppb.Now(),
		Secret:    clientSecret[0:10] + strings.Repeat("*", 32-10),
	}

	db.Info.Credentials = append(db.Info.Credentials, c)

	if _, err := s.dr.Update(ctx, db); err != nil {
		return "", "", err
	}

	return formatClientID(certificateID), clientSecret, nil
}
