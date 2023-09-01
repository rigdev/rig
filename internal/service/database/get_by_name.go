package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Service) GetByName(ctx context.Context, name string) (*database.Database, error) {
	db, err := s.dr.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	dbID := uuid.UUID(db.GetDatabaseId())

	switch db.GetType() {
	case database.Type_TYPE_MONGO:
		names, err := s.mongo.Database(formatDatabaseID(dbID)).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			return nil, err
		}
		db.Info.Tables = make([]*database.Table, len(names))
		for index, name := range names {
			db.Info.Tables[index] = &database.Table{Name: name}
		}
	case database.Type_TYPE_POSTGRES:
		res, err := s.postgres.Query("SELECT table_name FROM information_schema.tables")
		if err != nil {
			return nil, err
		}
		db.Info.Tables = []*database.Table{}
		for res.Next() {
			var table string
			if err := res.Scan(&table); err != nil {
				return nil, err
			}
			db.Info.Tables = append(db.Info.Tables, &database.Table{Name: table})
		}

	default:
		return nil, errors.InvalidArgumentErrorf("invalid database type: %v", db.GetType())
	}

	return db, nil
}
