package schema

import (
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

type Database struct {
	DatabaseID   uuid.UUID     `bson:"database_id" json:"database_id"`
	ProjectID    uuid.UUID     `bson:"project_id" json:"project_id"`
	Name         string        `bson:"name" json:"name"`
	DatabaseType database.Type `bson:"database_type" json:"database_type"`
	Data         []byte        `bson:"data" json:"data"`
}

func (d Database) ToProto() (*database.Database, error) {
	p := &database.Info{}
	if err := proto.Unmarshal(d.Data, p); err != nil {
		return nil, err
	}
	return &database.Database{
		DatabaseId: d.DatabaseID.String(),
		Name:       d.Name,
		Info:       p,
		Type:       d.DatabaseType,
	}, nil
}

func DatabaseFromProto(projectID uuid.UUID, d *database.Database) (Database, error) {
	bs, err := proto.Marshal(d.Info)
	if err != nil {
		return Database{}, err
	}

	return Database{
		DatabaseID:   uuid.UUID(d.GetDatabaseId()),
		ProjectID:    projectID,
		Name:         d.GetName(),
		DatabaseType: d.Type,
		Data:         bs,
	}, nil
}
