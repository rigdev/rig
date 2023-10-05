package schema

import (
	"github.com/rigdev/rig-go-api/api/v1/project"
	"google.golang.org/protobuf/proto"
)

type Project struct {
	ProjectID string `bson:"project_id" json:"project_id"`
	Name      string `bson:"name,omitempty" json:"name,omitempty"`
	Data      []byte `bson:"data,omitempty" json:"data,omitempty"`
}

func (p Project) ToProto() (*project.Project, error) {
	pr := &project.Project{}
	if err := proto.Unmarshal(p.Data, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func ProjectFromProto(p *project.Project) (Project, error) {
	bs, err := proto.Marshal(p)
	if err != nil {
		return Project{}, err
	}

	return Project{
		ProjectID: p.GetProjectId(),
		Name:      p.GetName(),
		Data:      bs,
	}, nil
}
