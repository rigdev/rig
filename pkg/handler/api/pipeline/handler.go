package pipeline

import (
	"github.com/go-logr/logr"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline/pipelineconnect"
	"github.com/rigdev/rig/pkg/service/objectstatus"
	svcpipeline "github.com/rigdev/rig/pkg/service/pipeline"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewHandler(
	pipeline svcpipeline.Service,
	objectstatus objectstatus.Service,
	scheme *runtime.Scheme,
	logger logr.Logger,
) pipelineconnect.ServiceHandler {
	return &handler{
		pipeline:     pipeline,
		objectstatus: objectstatus,
		scheme:       scheme,
		logger:       logger,
	}
}

type handler struct {
	pipeline     svcpipeline.Service
	objectstatus objectstatus.Service
	scheme       *runtime.Scheme
	logger       logr.Logger
}
