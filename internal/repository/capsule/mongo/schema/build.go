package schema

import (
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

type Build struct {
	ProjectID uuid.UUID `bson:"project_id" json:"project_id"`
	CapsuleID string    `bson:"capsule_id" json:"capsule_id"`
	BuildID   string    `bson:"build_id" json:"build_id"`
	Data      []byte    `bson:"data,omitempty" json:"data,omitempty"`
}

func (b Build) ToProto() (*capsule.Build, error) {
	p := &capsule.Build{}
	if err := proto.Unmarshal(b.Data, p); err != nil {
		return nil, err
	}

	return p, nil
}

func BuildFromProto(projectID uuid.UUID, capsuleID string, b *capsule.Build) (Build, error) {
	bs, err := proto.Marshal(b)
	if err != nil {
		return Build{}, err
	}

	return Build{
		ProjectID: projectID,
		CapsuleID: capsuleID,
		BuildID:   b.GetBuildId(),
		Data:      bs,
	}, nil
}
