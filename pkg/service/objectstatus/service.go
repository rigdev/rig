package objectstatus

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/go-logr/logr"
	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/pipeline"
	svc_pipeline "github.com/rigdev/rig/pkg/service/pipeline"
	"github.com/rigdev/rig/pkg/uuid"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type Service interface {
	// TODO: Adopt iterators.
	Watch(ctx context.Context, namespace string, c chan<- *apipipeline.ObjectStatusChange) error

	CapsulesInitialized()
	RegisterCapsule(namespace string, capsule string)
	UnregisterCapsule(namespace string, capsule string)

	UpdateStatus(namespace string, capsule string, pluginID uuid.UUID, change *apiplugin.ObjectStatusChange)
}

func NewService(
	cfg *v1alpha1.OperatorConfig,
	pipeline svc_pipeline.Service,
	logger logr.Logger,
) Service {
	s := &service{
		cfg:      cfg,
		logger:   logger,
		capsules: map[string]map[string]*capsuleCache{},
		pipeline: pipeline,
	}

	return s
}

type service struct {
	cfg         *v1alpha1.OperatorConfig
	logger      logr.Logger
	pipeline    svc_pipeline.Service
	lock        sync.RWMutex
	capsules    map[string]map[string]*capsuleCache
	watchers    []*watcher
	initialized bool
}

func (s *service) runForCapsule(ctx context.Context, namespace string, c *capsuleCache) {
	p := s.pipeline.GetDefaultPipeline()

	for _, step := range p.Steps() {
		for _, pluginID := range step.PluginIDs() {
			c.plugins[pluginID] = false
		}
		go s.runStepForCapsule(ctx, namespace, c.capsule, step)
	}
}

func (s *service) runStepForCapsule(
	ctx context.Context,
	namespace string,
	capsule string,
	step pipeline.Step[pipeline.CapsuleRequest],
) {
	for {
		if err := step.WatchObjectStatus(ctx, namespace, capsule, s); err == context.Canceled {
			return
		} else if err != nil {
			s.logger.Error(err, "error getting object status")
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *service) Watch(ctx context.Context, namespace string, c chan<- *apipipeline.ObjectStatusChange) error {
	// Important that the entire state (and all following events) are communicated from this point on.
	// Keep lock while building up and enqueuing a snapshot, and then start tailing.

	s.lock.Lock()
	w := newWatcher(namespace, c)
	s.watchers = append(s.watchers, w)

	defer func() {
		s.lock.Lock()
		s.watchers = slices.DeleteFunc(s.watchers, func(t *watcher) bool { return t == w })
		s.lock.Unlock()
	}()

	// Keep lock (and unlock in go-routine) when all statuses are read.
	go func() {
		s.readStatusForNamespace(namespace, w)
		s.sendCheckpoint(namespace, []*watcher{w})
		s.lock.Unlock()
	}()

	return w.run(ctx)
}

func (s *service) readStatusForNamespace(namespace string, w *watcher) []*apipipeline.ObjectStatusChange {
	capsules := s.capsules[namespace]

	var res []*apipipeline.ObjectStatusChange
	for _, capsule := range capsules {
		for _, s := range capsule.getStatuses() {
			if w.namespace == namespace {
				w.pushChange(s)
			}
		}
	}
	return res
}

func (s *service) CapsulesInitialized() {
	// Initialized!
	s.lock.Lock()
	defer s.lock.Unlock()
	s.initialized = true
	for namespace := range s.capsules {
		s.sendCheckpoint(namespace, s.watchers)
	}
}

func (s *service) RegisterCapsule(namespace string, capsule string) {
	if s.cfg.EnableObjectStatusCache == nil || !*s.cfg.EnableObjectStatusCache {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	cs, ok := s.capsules[namespace]
	if !ok {
		cs = map[string]*capsuleCache{}
		s.capsules[namespace] = cs
	}

	if _, ok := cs[capsule]; !ok {
		ctx, cancel := context.WithCancel(context.Background())
		c := &capsuleCache{
			plugins: map[uuid.UUID]bool{},
			capsule: capsule,
			cancel:  cancel,
			objects: map[pipeline.ObjectKey]*objectCache{},
		}
		cs[capsule] = c

		s.runForCapsule(ctx, namespace, c)
	}
}

func (s *service) UnregisterCapsule(namespace string, capsule string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ns, ok := s.capsules[namespace]
	if !ok {
		return
	}

	cs, ok := ns[capsule]
	if !ok {
		return
	}

	cs.cancel()

	delete(ns, capsule)

	// TODO: "delete" status event. We kind of want resource-events to out-live a capsule. Should we
	// delay cleanup until all objects are deleted? That could probably be kind of nice to do.

	if len(ns) == 0 {
		delete(s.capsules, namespace)
	}
}

func (s *service) UpdateStatus(
	namespace string,
	capsule string,
	pluginID uuid.UUID,
	change *apiplugin.ObjectStatusChange,
) {
	c := s.getCapsule(namespace, capsule)
	if c == nil {
		return
	}

	keys := c.update(pluginID, change)

	s.lock.RLock()
	for _, key := range keys {
		change := &apipipeline.ObjectStatusChange{
			Capsule: capsule,
		}
		if o, ok := c.objects[key]; ok {
			change.Change = &apipipeline.ObjectStatusChange_Updated{
				Updated: o.getStatus(),
			}
		} else {
			change.Change = &apipipeline.ObjectStatusChange_Deleted{
				Deleted: objectRefFromObjectKey(key),
			}
		}
		for _, w := range s.watchers {
			if w.namespace == namespace {
				w.pushChange(change)
			}
		}
	}

	if change.GetCheckpoint() != nil {
		s.sendCheckpoint(namespace, s.watchers)
	}

	s.lock.RUnlock()
}

func (s *service) sendCheckpoint(namespace string, watchers []*watcher) {
	if !s.initialized {
		return
	}

	for _, c := range s.capsules[namespace] {
		for _, initialized := range c.plugins {
			if !initialized {
				return
			}
		}
	}

	for _, w := range watchers {
		if w.namespace == namespace {
			w.checkpoint()
		}
	}
}

func (s *service) getCapsule(namespace string, capsule string) *capsuleCache {
	s.lock.RLock()
	defer s.lock.RUnlock()

	cs, ok := s.capsules[namespace]
	if !ok {
		return nil
	}

	c, ok := cs[capsule]
	if !ok {
		return nil
	}

	return c
}

type capsuleCache struct {
	// This property is owned by the service.
	plugins map[uuid.UUID]bool

	lock    sync.RWMutex
	capsule string
	objects map[pipeline.ObjectKey]*objectCache
	cancel  context.CancelFunc
}

func (c *capsuleCache) getStatuses() []*apipipeline.ObjectStatusChange {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var res []*apipipeline.ObjectStatusChange
	for _, object := range c.objects {
		res = append(res, &apipipeline.ObjectStatusChange{
			Capsule: c.capsule,
			Change: &apipipeline.ObjectStatusChange_Updated{
				Updated: object.getStatus(),
			},
		})
	}
	return res
}

func (c *capsuleCache) update(pluginID uuid.UUID, change *apiplugin.ObjectStatusChange) []pipeline.ObjectKey {
	c.lock.Lock()
	defer c.lock.Unlock()

	var keys []pipeline.ObjectKey
	switch v := change.GetChange().(type) {
	case *apiplugin.ObjectStatusChange_AllObjects_:
		// TODO: This is a hard reset - no need to update all objects.
		updated := map[pipeline.ObjectKey]struct{}{}
		for key, oc := range c.objects {
			if _, ok := oc.statuses[pluginID]; ok {
				updated[key] = struct{}{}
			}
		}

		for key := range updated {
			delete(c.objects[key].statuses, pluginID)
			if len(c.objects[key].statuses) == 0 {
				delete(c.objects, key)
			}
		}

		for _, obj := range v.AllObjects.GetObjects() {
			key := objectKeyFromObjectRef(obj.GetObjectRef())

			if c.objects[key] == nil {
				c.objects[key] = &objectCache{
					statuses: map[uuid.UUID]*apipipeline.ObjectStatus{},
				}
			}

			c.objects[key].statuses[pluginID] = obj
			updated[key] = struct{}{}
		}

		for key := range updated {
			keys = append(keys, key)
		}

	case *apiplugin.ObjectStatusChange_Updated:
		key := objectKeyFromObjectRef(v.Updated.GetObjectRef())

		if c.objects[key] == nil {
			c.objects[key] = &objectCache{
				statuses: map[uuid.UUID]*apipipeline.ObjectStatus{},
			}
		}

		current := c.objects[key].statuses[pluginID]
		if !proto.Equal(current, v.Updated) {
			c.objects[key].statuses[pluginID] = v.Updated
			keys = append(keys, key)
		}

	case *apiplugin.ObjectStatusChange_Deleted:
		key := objectKeyFromObjectRef(v.Deleted)

		if _, ok := c.objects[key].statuses[pluginID]; ok {
			delete(c.objects[key].statuses, pluginID)
			keys = append(keys, key)

			if len(c.objects[key].statuses) == 0 {
				delete(c.objects, key)
			}
		}

	case *apiplugin.ObjectStatusChange_Checkpoint_:
		c.plugins[pluginID] = true
	}

	return keys
}

type objectCache struct {
	statuses map[uuid.UUID]*apipipeline.ObjectStatus
}

func (o objectCache) getStatus() *apipipeline.ObjectStatus {
	keys := maps.Keys(o.statuses)
	slices.Sort(keys)
	res := &apipipeline.ObjectStatus{}
	for _, key := range keys {
		proto.Merge(res, o.statuses[key])
	}
	return res
}

func objectKeyFromObjectRef(ref *apipipeline.ObjectRef) pipeline.ObjectKey {
	return pipeline.ObjectKey{
		GroupVersionKind: schema.GroupVersionKind{
			Group:   ref.GetGvk().GetGroup(),
			Version: ref.GetGvk().GetVersion(),
			Kind:    ref.GetGvk().GetKind(),
		},
		ObjectKey: types.NamespacedName{
			Namespace: ref.GetNamespace(),
			Name:      ref.GetName(),
		},
	}
}

func objectRefFromObjectKey(key pipeline.ObjectKey) *apipipeline.ObjectRef {
	return &apipipeline.ObjectRef{
		Gvk: &apipipeline.GVK{
			Group:   key.Group,
			Version: key.Version,
			Kind:    key.Kind,
		},
		Name:      key.Name,
		Namespace: key.Namespace,
	}
}
