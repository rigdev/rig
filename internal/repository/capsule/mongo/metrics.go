package mongo

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/internal/repository/capsule/mongo/schema"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"go.mongodb.org/mongo-driver/bson"
)

func (c *MongoRepository) CreateMetrics(ctx context.Context, metrics *capsule.InstanceMetrics) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	m, err := schema.MetricFromProto(projectID, metrics)
	if err != nil {
		return err
	}

	if _, err := c.MetricsCol.InsertOne(ctx, m); err != nil {
		return fmt.Errorf("could not insert metric: %w", err)
	}

	return nil
}

func (c *MongoRepository) ListMetrics(
	ctx context.Context,
	pagination *model.Pagination,
) (iterator.Iterator[*capsule.InstanceMetrics], error) {
	return c.GetInstanceMetrics(ctx, pagination, "", "")
}

func (c *MongoRepository) GetMetrics(
	ctx context.Context,
	pagination *model.Pagination,
	capsuleID string,
) (iterator.Iterator[*capsule.InstanceMetrics], error) {
	return c.GetInstanceMetrics(ctx, pagination, capsuleID, "")
}

func (c *MongoRepository) GetInstanceMetrics(
	ctx context.Context,
	pagination *model.Pagination,
	capsuleID string,
	instanceID string,
) (iterator.Iterator[*capsule.InstanceMetrics], error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"project_id": projectID}
	if capsuleID != "" {
		filter["capsule_id"] = capsuleID
	}
	if instanceID != "" {
		filter["instance_id"] = instanceID
	}

	cursor, err := c.MetricsCol.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, fmt.Errorf("could not do find capsule metrics query: %w", err)
	}

	p := iterator.NewProducer[*capsule.InstanceMetrics]()
	go func() {
		defer p.Done()
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var m schema.CapsuleMetric
			if err := cursor.Decode(&m); err != nil {
				p.Error(fmt.Errorf("could not decode metric from cursor: %w", err))
				return
			}

			im, err := m.ToProto()
			if err != nil {
				p.Error(err)
				return
			}

			if err := p.Value(im); err != nil {
				p.Error(err)
				return
			}
		}
	}()

	return p, nil
}
