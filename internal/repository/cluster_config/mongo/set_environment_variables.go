package mongo

import (
	"context"

	"github.com/rigdev/rig/pkg/auth"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *MongoRepository) SetEnvironmentVariables(ctx context.Context, capsuleName string, envs map[string]string) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	if err := r.CapsuleConfigCol.FindOneAndUpdate(
		ctx,
		bson.M{
			"project_id": projectID,
			"name":       capsuleName,
		},
		bson.M{
			"$set": bson.M{
				"environmentVariables": envs,
			},
		},
	).Err(); err != nil {
		return err
	}

	return nil
}
