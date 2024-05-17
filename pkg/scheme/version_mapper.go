package scheme

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VersionMapper interface {
	FromGroupKind(gk schema.GroupKind) (schema.GroupVersionKind, error)
}

type clientVersionMapper struct {
	sync.Mutex

	cc       client.Client
	versions map[schema.GroupKind]schema.GroupVersionKind
}

func NewVersionMapper(cc client.Client) VersionMapper {
	return &clientVersionMapper{
		cc:       cc,
		versions: map[schema.GroupKind]schema.GroupVersionKind{},
	}
}

func (vm *clientVersionMapper) FromGroupKind(gk schema.GroupKind) (schema.GroupVersionKind, error) {
	vm.Lock()
	if gvk, ok := vm.versions[gk]; ok {
		vm.Unlock()
		return gvk, nil
	}
	vm.Unlock()

	res, err := vm.cc.RESTMapper().RESTMapping(gk)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	vm.Lock()
	vm.versions[gk] = res.GroupVersionKind
	vm.Unlock()

	return res.GroupVersionKind, nil
}

type schemeVersionMapper struct {
	scheme *runtime.Scheme
}

func NewVersionMapperFromScheme(scheme *runtime.Scheme) VersionMapper {
	return &schemeVersionMapper{scheme: scheme}
}

func (vm schemeVersionMapper) FromGroupKind(gk schema.GroupKind) (schema.GroupVersionKind, error) {
	gvs := vm.scheme.PrioritizedVersionsForGroup(gk.Group)
	if len(gvs) == 0 {
		return schema.GroupVersionKind{}, fmt.Errorf("unknown group '%s'", gk.Group)
	}

	return gk.WithVersion(gvs[0].Version), nil
}
