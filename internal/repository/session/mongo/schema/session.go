package schema

import (
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

type Session struct {
	ProjectID uuid.UUID `bson:"project_id" json:"project_id"`
	UserID    uuid.UUID `bson:"user_id" json:"user_id"`
	SessionID uuid.UUID `bson:"session_id" json:"session_id"`
	ExpiresAt int64     `bson:"expires_at" json:"expires_at"`
	Data      []byte    `bson:"data,omitempty" json:"data,omitempty"`
}

func (p Session) ToProto() (*user.Session, error) {
	pr := &user.Session{}
	if err := proto.Unmarshal(p.Data, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func SessionFromProto(projectID, userID, sessionID uuid.UUID, p *user.Session) (Session, error) {
	bs, err := proto.Marshal(p)
	if err != nil {
		return Session{}, err
	}

	return Session{
		ProjectID: projectID,
		UserID:    userID,
		SessionID: sessionID,
		ExpiresAt: p.GetExpiresAt().AsTime().Unix(),
		Data:      bs,
	}, nil
}

func GetSessionIDFilter(projectID, userID, sessionID uuid.UUID) bson.M {
	return bson.M{
		"project_id": projectID,
		"user_id":    userID,
		"session_id": sessionID,
	}
}
