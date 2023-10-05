package schema

import (
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

type Group struct {
	ProjectID string    `bson:"project_id" json:"project_id"`
	GroupID   uuid.UUID `bson:"group_id" json:"group_id"`
	Name      string    `bson:"name" json:"name"`
	Search    []string  `bson:"search" json:"search"`
	Data      []byte    `bson:"data" json:"data"`
}

func (g Group) ToProto() (*group.Group, error) {
	gr := &group.Group{}
	if err := proto.Unmarshal(g.Data, gr); err != nil {
		return nil, err
	}

	return gr, nil
}

func GroupFromProto(projectID string, g *group.Group) (Group, error) {
	bs, err := proto.Marshal(g)
	if err != nil {
		return Group{}, err
	}

	return Group{
		ProjectID: projectID,
		GroupID:   uuid.UUID(g.GetGroupId()),
		Name:      g.GetName(),
		Search: []string{
			g.GetGroupId(),
			g.GetName(),
		},
		Data: bs,
	}, nil
}

func GetGroupIDFilter(projectID string, groupID string) bson.M {
	return bson.M{
		"project_id": projectID,
		"group_id":   groupID,
	}
}
