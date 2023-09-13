package utils

import (
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

var DefaultResources = &capsule.Resources{
	Requests: &capsule.ResourceList{
		CpuMillis:   200,
		MemoryBytes: 512_000_000,
	},
	Limits: &capsule.ResourceList{
		CpuMillis:   0,
		MemoryBytes: 0,
	},
}

func FeedDefaultResources(r *capsule.Resources) {
	if r.Requests == nil {
		r.Requests = &capsule.ResourceList{}
	}
	feedDefaultResourceList(r.Requests, DefaultResources.Requests)

	if r.Limits == nil {
		r.Limits = &capsule.ResourceList{}
	}
	feedDefaultResourceList(r.Limits, DefaultResources.Limits)
}

func feedDefaultResourceList(r, defaultR *capsule.ResourceList) {
	if r.CpuMillis == 0 {
		r.CpuMillis = defaultR.CpuMillis
	}
	if r.MemoryBytes == 0 {
		r.MemoryBytes = defaultR.MemoryBytes
	}
}
