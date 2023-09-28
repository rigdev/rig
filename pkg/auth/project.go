package auth

import (
	"context"

	"github.com/rigdev/rig/pkg/errors"
)

var RigProjectID = "rig"

type projectIDKeyType string

const _projectIDKey projectIDKeyType = "projectID"

func WithProjectID(ctx context.Context, projectID string) context.Context {
	return context.WithValue(ctx, _projectIDKey, projectID)
}

func GetProjectID(ctx context.Context) (string, error) {
	val, ok := ctx.Value(_projectIDKey).(string)
	if ok {
		return val, nil
	}

	return "", errors.PermissionDeniedErrorf("no project selected")
}
