package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/pkg/repository/project/mongo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *MongoRepository) Update(ctx context.Context, p *project.Project) (*project.Project, error) {
	p.UpdatedAt = timestamppb.Now()
	u, err := schema.ProjectFromProto(p)
	if err != nil {
		return nil, err
	}

	if err := r.ProjectCol.FindOneAndUpdate(
		ctx,
		bson.M{"project_id": p.GetProjectId()},
		bson.M{
			"$set": u,
		},
	).Decode(&u); err != nil {
		return nil, err
	}
	return u.ToProto()
}
