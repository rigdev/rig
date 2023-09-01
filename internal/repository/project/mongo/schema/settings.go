package schema

import (
	"github.com/rigdev/rig/pkg/uuid"
)

type Settings struct {
	ProjectID uuid.UUID `bson:"project_id" json:"project_id"`
	Name      string    `bson:"name,omitempty" json:"name,omitempty"`
	Data      []byte    `bson:"data,omitempty" json:"data,omitempty"`
}
