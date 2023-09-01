package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Service) Get(ctx context.Context, databaseID uuid.UUID) (*database.Database, error) {
	var (
		db  *database.Database
		err error
	)
	db, err = s.dr.Get(ctx, databaseID)
	if err != nil {
		return nil, err
	}
	switch db.GetType() {
	case database.Type_TYPE_MONGO:
		names, err := s.mongo.Database(formatDatabaseID(databaseID)).ListCollectionNames(ctx, bson.M{})
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
