package schema

import (
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

type Database struct {
	DatabaseID uuid.UUID         `bson:"database_id" json:"database_id"`
	ProjectID  uuid.UUID         `bson:"project_id" json:"project_id"`
	Name       string            `bson:"name" json:"name"`
	Tables     []*database.Table `bson:"tables" json:"tables"`
	ClientIds  []string          `bson:"client_ids" json:"client_ids"`
	Data       []byte            `bson:"data" json:"data"`
}

func (d Database) ToProto() (*database.Database, error) {
	p := &database.Database{}
	if err := proto.Unmarshal(d.Data, p); err != nil {
		return nil, err
	}

	p.Tables = d.Tables
	p.ClientIds = d.ClientIds

	return p, nil
}

func DatabaseFromProto(projectID uuid.UUID, d *database.Database) (Database, error) {
	bs, err := proto.Marshal(d)
	if err != nil {
		return Database{}, err
	}

	return Database{
		DatabaseID: uuid.UUID(d.GetDatabaseId()),
		ProjectID:  projectID,
		Name:       d.GetName(),
		Tables:     d.GetTables(),
		ClientIds:  d.GetClientIds(),
		Data:       bs,
	}, nil
}
