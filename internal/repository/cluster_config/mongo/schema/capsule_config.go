package schema

import (
	"encoding/json"

	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/uuid"
	v1 "k8s.io/api/core/v1"
)

type CapsuleConfig struct {
	ProjectID            string            `bson:"project_id" json:"project_id"`
	CapsuleID            string            `bson:"capsule_id" json:"capsule_id"`
	Data                 []byte            `bson:"data,omitempty" json:"data,omitempty"`
	EnvironmentVariables map[string]string `bson:"environmentVariables" json:"environmentVariables"`
	Files                []*v1.ConfigMap   `bson:"files" json:"files"`
	Secrets              []*v1.Secret      `bson:"secrets" json:"secrets"`
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
		CapsuleID:            p.GetName(),
		EnvironmentVariables: map[string]string{},
		Data:                 bs,
	}, nil
}
