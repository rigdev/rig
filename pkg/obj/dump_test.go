package obj

import (
	"testing"

	"github.com/rigdev/rig/pkg/scheme"
	v1 "k8s.io/api/core/v1"
)

func TestDump(_ *testing.T) {
	o := &v1.ServiceAccount{}
	Dump(o, scheme.New())
}
