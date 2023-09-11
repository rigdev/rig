package utils

import (
	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

var DefaultResources = &capsule.Resources{
	Requests: &capsule.ResourceList{
		Cpu:              200,
		Memory:           512_000,
		EphemeralStorage: 512_000,
	},
	Limits: &capsule.ResourceList{
		Cpu:              0,
		Memory:           0,
		EphemeralStorage: 0,
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
	if r.Cpu == 0 {
		r.Cpu = defaultR.Cpu
	}
	if r.Memory == 0 {
		r.Memory = defaultR.Memory
	}
	if r.EphemeralStorage == 0 {
		r.EphemeralStorage = defaultR.EphemeralStorage
	}
}
