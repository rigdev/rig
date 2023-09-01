package model

import (
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type Secret struct {
	ProjectID uuid.UUID `bson:"project_id" json:"project_id"`
	SecretID  uuid.UUID `bson:"secret_id" json:"user_id"`
	Secret    []byte    `bson:"secret,omitempty" json:"secret,omitempty"`
}

func GetSecretIDFilter(projectID, secretID uuid.UUID) bson.M {
	return bson.M{
		"project_id": projectID,
		"secret_id":  secretID,
	}
}
