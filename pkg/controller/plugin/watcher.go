package plugin

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectWatcher interface {
	Reschedule(deadline time.Time)
	WatchSecondaryByName(objectName string, objType client.Object, cb WatchCallback)
	WatchSecondaryByLabels(selector labels.Selector, objType client.Object, cb WatchCallback)
}

type CapsuleWatcher interface {
	// Watch a primary resource, by providing a capsule name and
	// the type of the object to watch for.
	// All objects which type matches `objType` and that belongs to the capsule
	// will be processed.
	WatchPrimary(ctx context.Context, objType client.Object, cb WatchCallback) error
}

type capsuleWatcher struct {
	lock        sync.Mutex
	ctx         context.Context
	w           *watcher
	namespace   string
	capsule     string
	initialized bool
	c           chan<- *apiplugin.ObjectStatusChange
	initial     []*apipipeline.ObjectStatus
}

func (w *capsuleWatcher) flush() {
	w.lock.Lock()
	defer w.lock.Unlock()
	if w.initialized {
		return
	}

	init := w.initial
	w.initial = nil
	select {
	case w.c <- &apiplugin.ObjectStatusChange{
		Change: &apiplugin.ObjectStatusChange_AllObjects_{
			AllObjects: &apiplugin.ObjectStatusChange_AllObjects{
				Objects: init,
			},
		},
	}:
	case <-w.ctx.Done():
	}
	select {
	case w.c <- &apiplugin.ObjectStatusChange{
		Change: &apiplugin.ObjectStatusChange_Checkpoint_{
			Checkpoint: &apiplugin.ObjectStatusChange_Checkpoint{},
		},
	}:
	case <-w.ctx.Done():
	}

	w.initialized = true
}

func (w *capsuleWatcher) updated(os *apipipeline.ObjectStatus) {
	w.lock.Lock()
	if !w.initialized {
		w.initial = append(w.initial, os)
		w.lock.Unlock()
		return
	}
	w.lock.Unlock()

	select {
	case w.c <- &apiplugin.ObjectStatusChange{
		Change: &apiplugin.ObjectStatusChange_Updated{
			Updated: os,
		},
	}:
	case <-w.ctx.Done():
	}
}

func (w *capsuleWatcher) deleted(or *apipipeline.ObjectRef) {
	// TODO: This is probably not a good idea - delay instead?.
	w.flush()
	select {
	case w.c <- &apiplugin.ObjectStatusChange{
		Change: &apiplugin.ObjectStatusChange_Deleted{
			Deleted: or,
		},
	}:
	case <-w.ctx.Done():
	}
}

func (w *capsuleWatcher) WatchPrimary(ctx context.Context, objType client.Object, cb WatchCallback) error {
	return w.w.watchPrimary(ctx, w.namespace, w.capsule, objType, w, cb)
}

type WatchCallback func(
	obj client.Object,
	events []*corev1.Event,
	objectWatcher ObjectWatcher,
) *apipipeline.ObjectStatusInfo

type Watcher interface {
	NewCapsuleWatcher(
		ctx context.Context,
		namespace string,
		capsule string,
		c chan<- *apiplugin.ObjectStatusChange) CapsuleWatcher
}

func NewWatcher(logger hclog.Logger, cc client.WithWatch) Watcher {
	return &watcher{
		objectWatchers: map[watcherKey]*objectWatcher{},
		cc:             cc,
		logger:         logger,
	}
}

type watcherKey struct {
	namespace string
	gvk       schema.GroupVersionKind
}

type watcher struct {
	objectWatchers map[watcherKey]*objectWatcher
	cc             client.WithWatch
	objectSyncing  sync.WaitGroup
	logger         hclog.Logger
	lock           sync.Mutex
}

func (w *watcher) NewCapsuleWatcher(
	ctx context.Context,
	namespace string,
	capsule string,
	c chan<- *apiplugin.ObjectStatusChange,
) CapsuleWatcher {
	return &capsuleWatcher{
		ctx:       ctx,
		w:         w,
		namespace: namespace,
		capsule:   capsule,
		c:         c,
	}
}

func (w *watcher) watchPrimary(
	ctx context.Context,
	namespace string,
	capsule string,
	objType client.Object,
	cw *capsuleWatcher,
	cb WatchCallback,
) error {
	gvks, _, err := w.cc.Scheme().ObjectKinds(objType)
	if err != nil {
		return err
	}

	key := watcherKey{
		namespace: namespace,
		gvk:       gvks[0],
	}

	f := &objectWatch{
		key: objectWatchKey{
			watcherKey: key,
			labelSelector: labels.Set{
				pipeline.LabelOwnedByCapsule: capsule,
			}.AsSelector(),
		},
		cb:          cb,
		cw:          cw,
		subWatchers: map[string]*objectWatch{},
	}

	w.startWatch(f, objType)

	go func() {
		w.objectSyncing.Wait()
		cw.flush()
	}()

	<-ctx.Done()

	w.stopWatch(f)

	return nil
}

func (w *watcher) startWatch(f *objectWatch, objType client.Object) {
	w.lock.Lock()

	ow, ok := w.objectWatchers[f.key.watcherKey]
	if !ok {
		ow = newObjectWatcher(w, f.key.watcherKey.namespace, objType, w.cc, w.logger)
		w.objectWatchers[f.key.watcherKey] = ow
	}

	ow.addFilter(f)

	w.lock.Unlock()
}

func (w *watcher) stopWatch(f *objectWatch) {
	w.lock.Lock()

	if ow, ok := w.objectWatchers[f.key.watcherKey]; ok {
		if ow.removeFilter(f) {
			ow.stop()
			delete(w.objectWatchers, f.key.watcherKey)
		}
	}

	sf := f.subWatchers
	f.subWatchers = nil
	w.lock.Unlock()

	for _, sf := range sf {
		w.stopWatch(sf)
	}
}

type objectWatchKey struct {
	watcherKey    watcherKey
	names         []string
	labelSelector labels.Selector
}

// TODO: This is not a very optimal way of computing this!
func (k objectWatchKey) id() string {
	return fmt.Sprint(k.watcherKey, strings.Join(k.names, ","), k.labelSelector)
}

func (k objectWatchKey) matches(obj client.Object) bool {
	ls := obj.GetLabels()
	if k.labelSelector != nil && !k.labelSelector.Matches(labels.Set(ls)) {
		return false
	}

	if len(k.names) != 0 && !slices.Contains(k.names, obj.GetName()) {
		return false
	}

	return true
}

type objectWatch struct {
	key         objectWatchKey
	parent      *apipipeline.ObjectRef
	cb          WatchCallback
	cw          *capsuleWatcher
	subWatchers map[string]*objectWatch
}

type eventListWatcher struct {
	ctx       context.Context
	cc        client.WithWatch
	namespace string
	fields    fields.Set
	logger    hclog.Logger
}

func (w *eventListWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	list := &corev1.EventList{}
	if err := w.cc.List(w.ctx, list, &client.ListOptions{
		Namespace:     w.namespace,
		FieldSelector: fields.SelectorFromSet(w.fields),
		Raw:           &options,
	}); err != nil {
		w.logger.Error("error getting events", "fields", w.fields, "error", err)
		return nil, err
	}

	return list, nil
}

func (w *eventListWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	list := &corev1.EventList{}
	wi, err := w.cc.Watch(w.ctx, list, &client.ListOptions{
		Namespace:     w.namespace,
		FieldSelector: fields.SelectorFromSet(w.fields),
		Raw:           &options,
	})
	if err != nil {
		w.logger.Error("error watching events", "fields", w.fields, "error", err)
	}
	return wi, err
}

type queueObj struct {
	deadline time.Time
	obj      client.Object
	index    int
}

type objectWatcher struct {
	w       *watcher
	ctx     context.Context
	cancel  context.CancelFunc
	gvkList schema.GroupVersionKind
	cc      client.WithWatch
	logger  hclog.Logger
	store   cache.Store
	ctrl    cache.Controller

	eventStore cache.Store
	eventCtrl  cache.Controller

	namespace string

	lock sync.Mutex

	objects map[string]*queueObj
	queue   *priorityHeap[*queueObj]

	filters map[*objectWatch]struct{}
}

func newObjectWatcher(
	w *watcher,
	namespace string,
	obj client.Object,
	cc client.WithWatch,
	logger hclog.Logger,
) *objectWatcher {
	gvks, _, err := cc.Scheme().ObjectKinds(obj)
	if err != nil {
		log.Fatal(err)
	}

	gvk := gvks[0]

	gvkList := gvks[0]
	gvkList.Kind += "List"

	ctx, cancel := context.WithCancel(context.Background())

	ow := &objectWatcher{
		w:         w,
		ctx:       ctx,
		cancel:    cancel,
		cc:        cc,
		gvkList:   gvkList,
		logger:    logger,
		namespace: namespace,
		filters:   map[*objectWatch]struct{}{},
		objects:   map[string]*queueObj{},
		queue:     newPriorityHeap(func(a, b *queueObj) bool { return a.deadline.Before(b.deadline) }),
	}

	ow.queue.SetIndex(func(qo *queueObj, i int) {
		qo.index = i
	})

	w.logger.Info("starting watcher", "gvk", ow.gvkList)

	store, ctrl := cache.NewInformerWithOptions(cache.InformerOptions{
		ListerWatcher: ow,
		ObjectType:    obj,
		Handler:       ow,
	})
	ow.store = store
	ow.ctrl = ctrl

	apiVersion, kind := gvk.ToAPIVersionAndKind()
	elw := &eventListWatcher{
		ctx:       ctx,
		cc:        cc,
		namespace: namespace,
		fields: fields.Set{
			"involvedObject.apiVersion": apiVersion,
			"involvedObject.kind":       kind,
		},
		logger: logger,
	}
	eventStore, eventCtrl := cache.NewInformerWithOptions(cache.InformerOptions{
		ListerWatcher: elw,
		ObjectType:    &corev1.Event{},
		Handler:       ow,
	})
	ow.eventStore = eventStore
	ow.eventCtrl = eventCtrl

	w.objectSyncing.Add(1)

	go ow.run(ctx)

	return ow
}

func (ow *objectWatcher) stop() {
	ow.logger.Info("stopping watcher", "gvk", ow.gvkList)
	ow.cancel()
}

func (ow *objectWatcher) addFilter(f *objectWatch) {
	ow.lock.Lock()
	defer ow.lock.Unlock()

	ow.logger.Info("adding filter")

	// Flush existing objects.
	objects := ow.store.List()

	ow.filters[f] = struct{}{}

	for _, obj := range objects {
		co, ok := obj.(client.Object)
		if !ok {
			ow.logger.Info("invalid object type")
			continue
		}

		ow.handleForFilter(co, f, false)
	}
}

func (ow *objectWatcher) removeFilter(f *objectWatch) bool {
	ow.lock.Lock()
	defer ow.lock.Unlock()

	ow.logger.Info("removing filter")

	delete(ow.filters, f)

	// Flush delete events objects.
	objects := ow.store.List()

	for _, obj := range objects {
		co, ok := obj.(client.Object)
		if !ok {
			ow.logger.Info("invalid object type")
			continue
		}

		ow.handleForFilter(co, f, true)
	}

	return len(ow.filters) == 0
}

func (ow *objectWatcher) run(ctx context.Context) {
	go ow.ctrl.Run(ctx.Done())
	go ow.eventCtrl.Run(ctx.Done())

	ow.logger.Info("waiting for sync", "namespace", ow.namespace, "gvk", ow.gvkList)
	success := cache.WaitForCacheSync(ctx.Done(), ow.ctrl.HasSynced, ow.eventCtrl.HasSynced)
	ow.logger.Info("sync is done", "namespace", ow.namespace, "gvk", ow.gvkList, "success", success)

	ow.w.objectSyncing.Done()

	timer := time.NewTicker(250 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		ow.lock.Lock()

		for ow.queue.Len() > 0 && ow.queue.Peek().deadline.Before(time.Now()) {
			p := ow.queue.Pop()
			for f := range ow.filters {
				ow.handleForFilter(p.obj, f, false)
			}
		}

		ow.lock.Unlock()
	}
}

func (ow *objectWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	list, err := ow.cc.Scheme().New(ow.gvkList)
	if err != nil {
		return nil, err
	}

	if err := ow.cc.List(ow.ctx, list.(client.ObjectList), &client.ListOptions{
		Namespace: ow.namespace,
		Raw:       &options,
	}); err != nil {
		if errors.IsUnimplemented(errors.FromK8sClient(err)) {
			ow.logger.Info("type not registered", "gvk", ow.gvkList, "error", err)
			return list, nil
		}

		ow.logger.Error("error getting object list", "gvk", ow.gvkList, "error", err)
		return nil, err
	}

	return list, nil
}

func (ow *objectWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	list, err := ow.cc.Scheme().New(ow.gvkList)
	if err != nil {
		return nil, err
	}

	wi, err := ow.cc.Watch(ow.ctx, list.(client.ObjectList), &client.ListOptions{
		Namespace: ow.namespace,
		Raw:       &options,
	})
	if err != nil {
		ow.logger.Error("error watching objects", "gvk", ow.gvkList, "error", err)
	}
	return wi, err
}

func (ow *objectWatcher) OnAdd(obj any, _ bool) {
	if e, ok := obj.(*corev1.Event); ok {
		key := cache.NewObjectName(e.InvolvedObject.Namespace, e.InvolvedObject.Name)
		item, exists, err := ow.store.GetByKey(key.String())
		if err != nil {
			ow.logger.Error("error getting object from event", "gvk", ow.gvkList, "error", err)
		}
		if !exists {
			return
		}

		obj = item
	}

	co, ok := obj.(client.Object)
	if !ok {
		ow.logger.Info("invalid object type")
		return
	}

	ow.logger.Info("object updated", "gvk", ow.gvkList, "name", co.GetName(), "namespace", co.GetNamespace())

	ow.lock.Lock()
	defer ow.lock.Unlock()

	if o, ok := ow.objects[co.GetName()]; ok {
		o.obj = co

		if o.index >= 0 {
			// Already scheduled to run in the future.
			deadline := time.Now().Add(time.Second)
			if o.deadline.After(deadline) {
				o.deadline = deadline
				ow.queue.Fix(o.index)
			}
			return
		}

		// Deadline is when we can process the next event for this object.
		deadline := o.deadline.Add(1 * time.Second)
		if deadline.After(time.Now()) {
			o.deadline = deadline
			ow.queue.Push(o)
			return
		}

		// All good, ready to run it.
		o.deadline = time.Now()
	} else {
		ow.objects[co.GetName()] = &queueObj{
			deadline: time.Now(),
			obj:      co,
			index:    -1,
		}
	}

	for f := range ow.filters {
		ow.handleForFilter(co, f, false)
	}
}

func (ow *objectWatcher) OnUpdate(_, newObj any) {
	ow.OnAdd(newObj, false)
}

func (ow *objectWatcher) OnDelete(obj any) {
	if e, ok := obj.(*corev1.Event); ok {
		key := cache.NewObjectName(e.InvolvedObject.Namespace, e.InvolvedObject.Name)
		item, exists, err := ow.store.GetByKey(key.String())
		if err != nil {
			ow.logger.Error("error getting object from event", "gvk", ow.gvkList, "error", err)
		}
		if !exists {
			return
		}

		// Event deleted, but object still exists. Run OnAdd.
		ow.OnAdd(item, false)
		return
	}

	co, ok := obj.(client.Object)
	if !ok {
		ow.logger.Info("invalid object type")
		return
	}

	ow.logger.Info("object deleted", "gvk", ow.gvkList, "name", co.GetName(), "namespace", co.GetNamespace())

	ow.lock.Lock()
	defer ow.lock.Unlock()

	if q, ok := ow.objects[co.GetName()]; ok {
		if q.index >= 0 {
			// Already scheduled to run in the future, stop it.
			ow.queue.Remove(q.index)
		}
		delete(ow.objects, co.GetName())
	}

	for f := range ow.filters {
		ow.handleForFilter(co, f, true)
	}
}

func (ow *objectWatcher) handleForFilter(co client.Object, f *objectWatch, remove bool) {
	if !f.key.matches(co) {
		return
	}

	res := objectWatcherResult{
		cc:        ow.cc,
		namespace: ow.namespace,
		watchers:  map[string]objectWatchCandidate{},
	}

	gvks, _, err := ow.cc.Scheme().ObjectKinds(co)
	if err != nil {
		log.Fatal(err)
	}

	ref := &apipipeline.ObjectRef{
		Gvk: &apipipeline.GVK{
			Group:   gvks[0].Group,
			Version: gvks[0].Version,
			Kind:    gvks[0].Kind,
		},
		Namespace: co.GetNamespace(),
		Name:      co.GetName(),
	}

	if remove {
		f.cw.deleted(ref)
	} else {
		var events []*corev1.Event
		for _, e := range ow.eventStore.List() {
			event := e.(*corev1.Event)
			if event.InvolvedObject.Name != co.GetName() {
				continue
			}

			if eventTimestamp(event).Before(co.GetCreationTimestamp().Time) {
				continue
			}

			events = append(events, event)
		}
		info := f.cb(co, events, &res)
		status := &apipipeline.ObjectStatus{
			ObjectRef: ref,
			Info:      info,
			CreatedAt: timestamppb.New(co.GetCreationTimestamp().Time),
			UpdatedAt: timestamppb.Now(),
			Parent:    f.parent,
		}

		if !co.GetDeletionTimestamp().IsZero() {
			status.DeletionAt = timestamppb.New(co.GetDeletionTimestamp().Time)
		}

		f.cw.updated(status)

		if !res.deadline.IsZero() {
			// Must exist in map!
			o := ow.objects[co.GetName()]
			if o.index >= 0 {
				if res.deadline.Before(o.deadline) {
					o.deadline = res.deadline
					ow.queue.Fix(o.index)
				}
			} else {
				o.deadline = res.deadline
				ow.queue.Push(o)
			}
		}
	}

	for key, w := range res.watchers {
		if _, ok := f.subWatchers[key]; !ok {
			sf := &objectWatch{
				key:         w.key,
				parent:      ref,
				cb:          w.cb,
				cw:          f.cw,
				subWatchers: map[string]*objectWatch{},
			}

			f.subWatchers[key] = sf
			go ow.w.startWatch(sf, w.objType)
		}
	}

	// Check if not generated anymore
	for key, sw := range f.subWatchers {
		if _, ok := res.watchers[key]; !ok {
			delete(f.subWatchers, key)
			go ow.w.stopWatch(sw)
		}
	}
}

func eventTimestamp(event *corev1.Event) time.Time {
	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp.Time
	}

	if !event.EventTime.IsZero() {
		return event.EventTime.Time
	}

	if !event.FirstTimestamp.IsZero() {
		return event.FirstTimestamp.Time
	}

	return event.CreationTimestamp.Time
}

type objectWatcherResult struct {
	cc        client.WithWatch
	namespace string
	watchers  map[string]objectWatchCandidate
	deadline  time.Time
}

func (r *objectWatcherResult) watchObject(key objectWatchKey, objType client.Object, cb WatchCallback) {
	gvks, _, err := r.cc.Scheme().ObjectKinds(objType)
	if err != nil {
		// TODO!
		log.Fatal(err)
	}

	key.watcherKey = watcherKey{
		namespace: r.namespace,
		gvk:       gvks[0],
	}

	r.watchers[key.id()] = objectWatchCandidate{
		key:     key,
		objType: objType,
		cb:      cb,
	}
}

func (r *objectWatcherResult) WatchSecondaryByName(objectName string, objType client.Object, cb WatchCallback) {
	r.watchObject(
		objectWatchKey{
			names: []string{objectName},
		},
		objType,
		cb,
	)
}

func (r *objectWatcherResult) WatchSecondaryByLabels(
	objectLabels labels.Selector, objType client.Object, cb WatchCallback,
) {
	r.watchObject(
		objectWatchKey{
			labelSelector: objectLabels,
		},
		objType,
		cb,
	)
}

func (r *objectWatcherResult) Reschedule(deadline time.Time) {
	if r.deadline.IsZero() || deadline.Before(r.deadline) {
		r.deadline = deadline
	}
}

type objectWatchCandidate struct {
	key     objectWatchKey
	objType client.Object
	cb      WatchCallback
}
