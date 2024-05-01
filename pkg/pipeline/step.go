package pipeline

import (
	"context"

	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/uuid"
)

type ObjectStatusCallback interface {
	UpdateStatus(namespace string, capsule string, pluginID uuid.UUID, change *apiplugin.ObjectStatusChange)
}

type Step[T Request] interface {
	Apply(ctx context.Context, req T) error
	WatchObjectStatus(ctx context.Context, namespace, capsule string, callback ObjectStatusCallback) error
}
