package log

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func New(devMode bool) logr.Logger {
	return zap.New(zap.UseDevMode(devMode))
}
