package pipeline

import (
	"context"

	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectStatusCallback interface {
	UpdateStatus(namespace string, capsule string, pluginID uuid.UUID, change *apiplugin.ObjectStatusChange)
}

type PipelineOptions struct {
	AdditionalObjects []client.Object
}

type Step[T Request] interface {
	Apply(ctx context.Context, req T, opts PipelineOptions) error
	WatchObjectStatus(ctx context.Context, namespace, capsule string, callback ObjectStatusCallback) error
	PluginIDs() []uuid.UUID
}
