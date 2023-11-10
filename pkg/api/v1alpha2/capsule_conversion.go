package v1alpha2

import "sigs.k8s.io/controller-runtime/pkg/conversion"

var _ conversion.Hub = &Capsule{}

func (c *Capsule) Hub() {}
