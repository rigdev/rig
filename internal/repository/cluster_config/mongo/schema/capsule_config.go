package schema

import (
	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

type CapsuleConfig struct {
	ProjectID string `bson:"project_id" json:"project_id"`
	Name      string `bson:"name" json:"name"`
	Data      []byte `bson:"data,omitempty" json:"data,omitempty"`
}

func (c CapsuleConfig) ToProto() (*capsule.Config, error) {
	p := &capsule.Config{}
	if err := proto.Unmarshal(c.Data, p); err != nil {
		return nil, err
	}

	return p, nil
}

func CapsuleConfigFromProto(projectID uuid.UUID, p *capsule.Config) (CapsuleConfig, error) {
	bs, err := proto.Marshal(p)
	if err != nil {
		return CapsuleConfig{}, err
	}

	return CapsuleConfig{
		ProjectID: projectID.String(),
		Name:      p.GetName(),
		Data:      bs,
	}, nil
}
