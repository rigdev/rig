package mongo

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// DeleteBatch deletes a batch of user by id, email or username.
func (r *MongoRepository) DeleteMany(ctx context.Context, userBatch []*model.UserIdentifier) (uint64, error) {
	projectId, err := auth.GetProjectID(ctx)
	if err != nil {
		return 0, err
	}

	var ids []string
	var emails []string
	var usernames []string
	var phoneNumbers []string
	for _, user := range userBatch {
		if user == nil {
			return 0, errors.InvalidArgumentErrorf("user cannot be nil")
		}
		if user.GetEmail() != "" {
			emails = append(emails, user.GetEmail())
		} else if user.GetUsername() != "" {
			usernames = append(usernames, user.GetUsername())
		} else if user.GetPhoneNumber() != "" {
			phoneNumbers = append(phoneNumbers, user.GetPhoneNumber())
		}
	}
	combinedFilter := bson.A{}
	if len(ids) > 0 {
		combinedFilter = append(combinedFilter, bson.M{"_id": bson.D{{Key: "$in", Value: ids}}})
	}
	if len(emails) > 0 {
		combinedFilter = append(combinedFilter, bson.M{"email": bson.D{{Key: "$in", Value: emails}}})
	}
	if len(usernames) > 0 {
		combinedFilter = append(combinedFilter, bson.M{"username": bson.D{{Key: "$in", Value: usernames}}})
	}
	if len(phoneNumbers) > 0 {
		combinedFilter = append(combinedFilter, bson.M{"phone": bson.D{{Key: "$in", Value: phoneNumbers}}})
	}
	res, err := r.UsersCol.DeleteMany(ctx, bson.D{{Key: "$or", Value: combinedFilter}, {Key: "project_Id", Value: projectId}})
	if err != nil {
		return 0, err
	}

	return uint64(res.DeletedCount), nil
}
