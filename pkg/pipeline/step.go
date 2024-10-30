package pipeline

import (
	"context"

	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectStatusCallback interface {
	UpdateStatus(namespace string, capsule string, pluginID uuid.UUID, change *apiplugin.ObjectStatusChange)
}

type Options struct {
	AdditionalObjects []client.Object
}

type Step[T Request] interface {
	Name() string
	Apply(ctx context.Context, req T, opts Options) error
	WatchObjectStatus(ctx context.Context, capsule *v1alpha2.Capsule, callback ObjectStatusCallback) error
	ComputeConfig(ctx context.Context, req T) StepConfigResult
	PluginIDs() []uuid.UUID
}
