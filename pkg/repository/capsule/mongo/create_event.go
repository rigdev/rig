package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
)

func (m *MongoRepository) CreateEvent(ctx context.Context, capsuleID string, e *capsule.Event) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	bp, err := schema.EventFromProto(projectID, capsuleID, e)
	if err != nil {
		return err
	}

	if _, err := m.CapsuleEventCol.InsertOne(ctx, bp); err != nil {
		return err
	}

	return nil
}
