package schema

import (
	"encoding/json"

	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/uuid"
)

type CapsuleConfig struct {
	ProjectID            string            `bson:"project_id" json:"project_id"`
	Name                 string            `bson:"name" json:"name"`
	EnvironmentVariables map[string]string `bson:"environmentVariables" json:"environmentVariables"`
	Data                 []byte            `bson:"data,omitempty" json:"data,omitempty"`
}

func (c CapsuleConfig) ToAPI() (*v1alpha1.Capsule, error) {
	p := &v1alpha1.Capsule{}
	if err := json.Unmarshal(c.Data, p); err != nil {
		return nil, err
	}

	return p, nil
}

func CapsuleConfigFromAPI(projectID uuid.UUID, p *v1alpha1.Capsule) (CapsuleConfig, error) {
	bs, err := json.Marshal(p)
	if err != nil {
		return CapsuleConfig{}, err
	}

	return CapsuleConfig{
		ProjectID:            projectID.String(),
		Name:                 p.GetName(),
		EnvironmentVariables: map[string]string{},
		Data:                 bs,
	}, nil
}
