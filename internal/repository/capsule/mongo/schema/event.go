package schema

import (
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

type Event struct {
	ProjectID uuid.UUID `bson:"project_id" json:"project_id"`
	CapsuleID string    `bson:"capsule_id" json:"capsule_id"`
	RolloutID uint64    `bson:"rollout_id" json:"rollout_id"`
	Data      []byte    `bson:"data,omitempty" json:"data,omitempty"`
}

func (e Event) ToProto() (*capsule.Event, error) {
	p := &capsule.Event{}
	if err := proto.Unmarshal(e.Data, p); err != nil {
		return nil, err
	}

	return p, nil
}

func EventFromProto(projectID uuid.UUID, capsuleID string, e *capsule.Event) (Event, error) {
	bs, err := proto.Marshal(e)
	if err != nil {
		return Event{}, err
	}

	return Event{
		ProjectID: projectID,
		CapsuleID: capsuleID,
		RolloutID: e.GetRolloutId(),
		Data:      bs,
	}, nil
}
