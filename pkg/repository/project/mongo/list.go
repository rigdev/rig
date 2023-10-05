package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/client/mongo"
	"github.com/rigdev/rig/pkg/repository/project/mongo/schema"
	"github.com/rigdev/rig/pkg/iterator"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *MongoRepository) List(ctx context.Context, pagination *model.Pagination, exclude []string) (iterator.Iterator[*project.Project], int64, error) {
	filter := bson.M{
		"project_id": bson.M{
			"$nin": exclude,
		},
	}

	count, err := r.ProjectCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := r.ProjectCol.Find(ctx, filter, mongo.SortOptions(pagination))
	if err != nil {
		return nil, 0, err
	}

	it := iterator.NewProducer[*project.Project]()
	go func() {
		defer it.Done()
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var p schema.Project
			if err := cursor.Decode(&p); err != nil {
				it.Error(err)
				return
			}
			e, err := p.ToProto()
			if err != nil {
				it.Error(err)
				return
			}
			if err := it.Value(e); err != nil {
				it.Error(err)
				return
			}
		}
	}()

	return it, count, nil
}
