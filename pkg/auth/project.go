package auth

import (
	"context"

	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

var RigProjectID = uuid.MustParse("c10c947b-91f1-41ea-96df-ea13ee68a7fc")

type projectIDKeyType string

const _projectIDKey projectIDKeyType = "projectID"

func WithProjectID(ctx context.Context, projectID uuid.UUID) context.Context {
	return context.WithValue(ctx, _projectIDKey, projectID)
}

func GetProjectID(ctx context.Context) (uuid.UUID, error) {
	val, ok := ctx.Value(_projectIDKey).(uuid.UUID)
	if ok {
		return val, nil
	}

	return uuid.Nil, errors.PermissionDeniedErrorf("no project selected")
}
