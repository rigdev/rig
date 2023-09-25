package auth

import (
	"context"

	"github.com/rigdev/rig/pkg/errors"
)

var RigProjectID = "c10c947b-91f1-41ea-96df-ea13ee68a7fc"

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
