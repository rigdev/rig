package schema

import (
	"time"

	"github.com/rigdev/rig/pkg/uuid"
)

type GroupMember struct {
	ProjectID uuid.UUID `bson:"project_id" json:"project_id"`
	GroupID   uuid.UUID `bson:"group_id" json:"group_id"`
	UserID    uuid.UUID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
