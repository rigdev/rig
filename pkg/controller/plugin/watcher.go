package plugin

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectWatcher interface {
	WatchSecondaryByName(objectName string, objType runtime.Object, cb WatchCallback)
}

type CapsuleWatcher interface {
	// Watch a primary resource, by providing a capsule name and
	// the type of the object to watch for.
	// All objects which type matches `objType` and that belongs to the capsule
	// will be processed.
	WatchPrimary(ctx context.Context, objType runtime.Object, cb WatchCallback) error
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

func (w *capsuleWatcher) WatchPrimary(ctx context.Context, objType runtime.Object, cb WatchCallback) error {
	return w.w.watchPrimary(ctx, w.namespace, w.capsule, objType, w, cb)
}

type WatchCallback func(obj runtime.Object, objectWatcher ObjectWatcher) *apipipeline.ObjectStatus

type Watcher interface {
	NewCapsuleWatcher(ctx context.Context, namespace, capsule string, c chan<- *apiplugin.ObjectStatusChange) CapsuleWatcher
}

func NewWatcher(logger hclog.Logger, cc client.WithWatch) Watcher {
	return &watcher{
		objectWatchers: map[watcherKey]*objectWatcher[runtime.Object]{},
		cc:             cc,
		logger:         logger,
	}
}

type watcherKey struct {
	namespace string
	gvk       schema.GroupVersionKind
}

type watcher struct {
	objectWatchers map[watcherKey]*objectWatcher[runtime.Object]
	cc             client.WithWatch
	objectSyncing  sync.WaitGroup
	logger         hclog.Logger
	lock           sync.Mutex
}

func (w *watcher) NewCapsuleWatcher(ctx context.Context, namespace, capsule string, c chan<- *apiplugin.ObjectStatusChange) CapsuleWatcher {
	return &capsuleWatcher{
		ctx:       ctx,
		w:         w,
		namespace: namespace,
		capsule:   capsule,
		c:         c,
	}
}

func (w *watcher) watchPrimary(ctx context.Context, namespace, capsule string, objType runtime.Object, cw *capsuleWatcher, cb WatchCallback) error {
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
			labels: labels.Set{
				pipeline.LabelOwnedByCapsule: capsule,
			},
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

func (w *watcher) startWatch(f *objectWatch, objType runtime.Object) {
	w.lock.Lock()

	ow, ok := w.objectWatchers[f.key.watcherKey]
	if !ok {
		ow = NewObjectWatcher(w, f.key.watcherKey.namespace, objType, w.cc, w.logger)
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
	watcherKey watcherKey
	names      []string
	labels     labels.Set
}

// TODO: This is not a very optimal way of computing this!
func (k objectWatchKey) id() string {
	return fmt.Sprint(k.watcherKey, strings.Join(k.names, ","), k.labels)
}

type objectWatch struct {
	key         objectWatchKey
	cb          WatchCallback
	cw          *capsuleWatcher
	subWatchers map[string]*objectWatch
}

func (f *objectWatchKey) matches(obj client.Object) bool {
	ls := obj.GetLabels()
	if !f.labels.AsSelector().Matches(labels.Set(ls)) {
		return false
	}

	if len(f.names) != 0 && !slices.Contains(f.names, obj.GetName()) {
		return false
	}

	return true
}

type objectWatcher[T runtime.Object] struct {
	w       *watcher
	ctx     context.Context
	cancel  context.CancelFunc
	gvkList schema.GroupVersionKind
	cc      client.WithWatch
	logger  hclog.Logger
	store   cache.Store
	ctrl    cache.Controller

	namespace string

	lock sync.Mutex

	filters map[*objectWatch]struct{}
}

type ToObjectStatus[T runtime.Object] func(obj T) *apipipeline.ObjectStatus

func NewObjectWatcher[T runtime.Object](w *watcher, namespace string, obj T, cc client.WithWatch, logger hclog.Logger) *objectWatcher[T] {
	gvks, _, err := cc.Scheme().ObjectKinds(obj)
	if err != nil {
		log.Fatal(err)
	}

	gvkList := gvks[0]
	gvkList.Kind += "List"

	ctx, cancel := context.WithCancel(context.Background())

	ow := &objectWatcher[T]{
		w:         w,
		ctx:       ctx,
		cancel:    cancel,
		cc:        cc,
		gvkList:   gvkList,
		logger:    logger,
		namespace: namespace,
		filters:   map[*objectWatch]struct{}{},
	}

	w.logger.Info("starting watcher", "gvk", ow.gvkList)

	store, ctrl := cache.NewInformer(ow, obj, 0, ow)
	ow.store = store
	ow.ctrl = ctrl

	w.objectSyncing.Add(1)

	go ow.run(ctx)

	return ow
}

func (ow *objectWatcher[T]) stop() {
	ow.logger.Info("stopping watcher", "gvk", ow.gvkList)
	ow.cancel()
}

func (ow *objectWatcher[T]) addFilter(f *objectWatch) {
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

func (w *objectWatcher[T]) removeFilter(f *objectWatch) bool {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.logger.Info("removing filter")

	delete(w.filters, f)
	return len(w.filters) == 0
}

func (w *objectWatcher[T]) run(ctx context.Context) {
	go w.ctrl.Run(ctx.Done())

	cache.WaitForCacheSync(ctx.Done(), w.ctrl.HasSynced)

	w.w.objectSyncing.Done()

	<-ctx.Done()
}

func (w *objectWatcher[T]) List(options metav1.ListOptions) (runtime.Object, error) {
	list, err := w.cc.Scheme().New(w.gvkList)
	if err != nil {
		return nil, err
	}

	if err := w.cc.List(w.ctx, list.(client.ObjectList), &client.ListOptions{
		Namespace: w.namespace,
		Raw:       &options,
	}); err != nil {
		return nil, err
	}

	return list, nil
}

func (w *objectWatcher[T]) Watch(options metav1.ListOptions) (watch.Interface, error) {
	list, err := w.cc.Scheme().New(w.gvkList)
	if err != nil {
		return nil, err
	}

	return w.cc.Watch(w.ctx, list.(client.ObjectList), &client.ListOptions{
		Namespace: w.namespace,
		Raw:       &options,
	})
}

func (ow *objectWatcher[T]) OnAdd(obj interface{}, isInInitialList bool) {
	co, ok := obj.(client.Object)
	if !ok {
		ow.logger.Info("invalid object type")
		return
	}

	ow.logger.Info("object updated", "gvk", ow.gvkList, "name", co.GetName())

	ow.lock.Lock()
	defer ow.lock.Unlock()
	for f := range ow.filters {
		ow.handleForFilter(co, f, false)
	}
}

func (w *objectWatcher[T]) OnUpdate(oldObj, newObj interface{}) {
	w.OnAdd(newObj, false)
}

func (ow *objectWatcher[T]) OnDelete(obj interface{}) {
	co, ok := obj.(client.Object)
	if !ok {
		ow.logger.Info("invalid object type")
		return
	}

	ow.logger.Info("object deleted", "gvk", ow.gvkList, "name", co.GetName())

	ow.lock.Lock()
	defer ow.lock.Unlock()
	for f := range ow.filters {
		ow.handleForFilter(co, f, true)
	}
}

func (ow *objectWatcher[T]) handleForFilter(co client.Object, f *objectWatch, remove bool) {
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
		os := f.cb(co, &res)
		os.ObjectRef = ref
		f.cw.updated(os)
	}

	for key, w := range res.watchers {
		if _, ok := f.subWatchers[key]; !ok {
			sf := &objectWatch{
				key:         w.key,
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

type objectWatcherResult struct {
	cc        client.WithWatch
	namespace string
	watchers  map[string]objectWatchCandidate
}

func (r *objectWatcherResult) WatchSecondaryByName(objectName string, objType runtime.Object, cb WatchCallback) {
	gvks, _, err := r.cc.Scheme().ObjectKinds(objType)
	if err != nil {
		// TODO!
		log.Fatal(err)
	}

	key := objectWatchKey{
		watcherKey: watcherKey{
			namespace: r.namespace,
			gvk:       gvks[0],
		},
		names: []string{objectName},
	}

	r.watchers[key.id()] = objectWatchCandidate{
		key:     key,
		objType: objType,
		cb:      cb,
	}
}

type objectWatchCandidate struct {
	key     objectWatchKey
	objType runtime.Object
	cb      WatchCallback
}
