package objectstatus

import (
	"context"
	"sync"

	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
)

type watcher struct {
	namespace string
	c         chan<- *apipipeline.ObjectStatusChange
	queue     []*apipipeline.ObjectStatusChange
	cond      *sync.Cond
}

func newWatcher(namespace string, c chan<- *apipipeline.ObjectStatusChange) *watcher {
	return &watcher{
		namespace: namespace,
		c:         c,
		cond:      sync.NewCond(&sync.Mutex{}),
	}
}

func (w *watcher) run(ctx context.Context) error {
	done := false
	go func() {
		<-ctx.Done()
		w.cond.L.Lock()
		done = true
		w.cond.Signal()
		w.cond.L.Unlock()
	}()

	for {
		w.cond.L.Lock()
		for {
			if done {
				w.cond.L.Unlock()
				return ctx.Err()
			}

			if len(w.queue) > 0 {
				break
			}

			w.cond.Wait()
		}
		element := w.queue[0]
		w.queue = w.queue[1:]
		w.cond.L.Unlock()

		select {
		case w.c <- element:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (w *watcher) pushChange(change *apipipeline.ObjectStatusChange) {
	w.cond.L.Lock()
	w.queue = append(w.queue, change)
	w.cond.Signal()
	w.cond.L.Unlock()
}
