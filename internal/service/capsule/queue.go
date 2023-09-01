package capsule

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

type QueueItem[T interface{}] struct {
	t        T
	deadline time.Time
}

type QueueHeap[T interface{}] []QueueItem[T]

func (qs QueueHeap[T]) Len() int {
	return len(qs)
}

func (qs QueueHeap[T]) Less(i, j int) bool {
	return qs[i].deadline.Before(qs[j].deadline)
}

func (qs QueueHeap[T]) Swap(i, j int) {
	qs[i], qs[j] = qs[j], qs[i]
}

func (qs *QueueHeap[T]) Pop() any {
	n := len(*qs) - 1
	t := (*qs)[n]
	*qs = (*qs)[:n]
	return t
}

func (qs QueueHeap[T]) Peek() QueueItem[T] {
	return qs[0]
}

func (qs *QueueHeap[T]) Push(t any) {
	*qs = append(*qs, t.(QueueItem[T]))
}

type Queue[T interface{}] struct {
	lock sync.Mutex
	is   *QueueHeap[T]
	c    chan struct{}
}

func NewQueue[T interface{}]() *Queue[T] {
	return &Queue[T]{
		c:  make(chan struct{}, 1),
		is: &QueueHeap[T]{},
	}
}

func (q *Queue[T]) AddJob(t T, deadline time.Time) {
	q.lock.Lock()
	heap.Push(q.is, QueueItem[T]{
		t:        t,
		deadline: deadline,
	})
	select {
	case q.c <- struct{}{}:
	default:
	}
	q.lock.Unlock()
}

func (q *Queue[T]) Next(ctx context.Context, now func() time.Time) (t T, err error) {
	var tot *time.Timer
	for err == nil {
		q.lock.Lock()
		d := time.Hour
		if q.is.Len() > 0 {
			qi := q.is.Peek()
			n := now()
			if !n.Before(qi.deadline) {
				heap.Pop(q.is)
				// Notify potential other listeners.
				select {
				case q.c <- struct{}{}:
				default:
				}
				q.lock.Unlock()
				return qi.t, nil
			}

			d = qi.deadline.Sub(n)
		}
		q.lock.Unlock()

		tot = time.NewTimer(d)
		select {
		case <-tot.C:
		case <-ctx.Done():
			if !tot.Stop() {
				<-tot.C
			}
			err = ctx.Err()
		case <-q.c:
			if !tot.Stop() {
				<-tot.C
			}
		}
	}

	return
}
