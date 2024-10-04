// nolint
package plugin

import (
	"container/heap"
)

// This implementation is copied from here:
// https://gotipplay.golang.org/p/d4M0QBkfmIr
// https://github.com/golang/go/issues/47632

// A priorityHeap is a min-heap backed by a slice.
type priorityHeap[E any] struct {
	s sliceHeap[E]
}

// newPriorityHeap constructs a new Heap with a comparison function.
func newPriorityHeap[E any](less func(E, E) bool) *priorityHeap[E] {
	return &priorityHeap[E]{sliceHeap[E]{less: less}}
}

// Push pushes an element onto the heap. The complexity is O(log n)
// where n = h.Len().
func (h *priorityHeap[E]) Push(elem E) {
	heap.Push(&h.s, elem)
}

// Pop removes and returns the minimum element (according to the less function)
// from the heap. Pop panics if the heap is empty.
// The complexity is O(log n) where n = h.Len().
func (h *priorityHeap[E]) Pop() E {
	return heap.Pop(&h.s).(E)
}

// Peek returns the minimum element (according to the less function) in the heap.
// Peek panics if the heap is empty.
// The complexity is O(1).
func (h *priorityHeap[E]) Peek() E {
	return h.s.s[0]
}

// Len returns the number of elements in the heap.
func (h *priorityHeap[E]) Len() int {
	return len(h.s.s)
}

// Slice returns the underlying slice.
// The slice is in heap order; the minimum value is at index 0.
// The heap retains the returned slice, so altering the slice may break
// the invariants and invalidate the heap.
func (h *priorityHeap[E]) Slice() []E {
	return h.s.s
}

// SetIndex specifies an optional function to be called
// when updating the position of any heap element within the slice,
// including during the element's initial Push.
//
// SetIndex must be called at most once, before any calls to Push.
//
// When an element is removed from the heap by Pop or Remove,
// the index function is called with the invalid index -1
// to signify that the element is no longer within the slice.
func (h *priorityHeap[E]) SetIndex(f func(E, int)) {
	h.s.setIndex = f
}

// Fix re-establishes the heap ordering
// after the element at index i has changed its value.
// Changing the value of the element at index i and then calling Fix
// is equivalent to, but less expensive than,
// calling h.Remove(i) followed by a Push of the new value.
// The complexity is O(log n) where n = h.Len().
// The index for use with Fix is recorded using the function passed to SetIndex.
func (h *priorityHeap[E]) Fix(i int) {
	heap.Fix(&h.s, i)
}

// Remove removes and returns the element at index i from the heap.
// The complexity is O(log n) where n = h.Len().
// The index for use with Remove is recorded using the function passed to SetIndex.
func (h *priorityHeap[E]) Remove(i int) E {
	return heap.Remove(&h.s, i).(E)
}

// sliceHeap just exists to use the existing heap.Interface as the
// implementation of Heap.
type sliceHeap[E any] struct {
	s        []E
	less     func(E, E) bool
	setIndex func(E, int)
}

func (s *sliceHeap[E]) Len() int { return len(s.s) }

func (s *sliceHeap[E]) Swap(i, j int) {
	s.s[i], s.s[j] = s.s[j], s.s[i]
	if s.setIndex != nil {
		s.setIndex(s.s[i], i)
		s.setIndex(s.s[j], j)
	}
}

func (s *sliceHeap[E]) Less(i, j int) bool {
	return s.less(s.s[i], s.s[j])
}

func (s *sliceHeap[E]) Push(x any) {
	s.s = append(s.s, x.(E))
	if s.setIndex != nil {
		s.setIndex(s.s[len(s.s)-1], len(s.s)-1)
	}
}

func (s *sliceHeap[E]) Pop() any {
	e := s.s[len(s.s)-1]
	if s.setIndex != nil {
		s.setIndex(e, -1)
	}
	s.s = s.s[:len(s.s)-1]
	return e
}

type Item struct {
	value    string
	priority int
	index    int
}

func (it *Item) Less(it1 *Item) bool {
	return it.priority > it1.priority
}

func (it *Item) setIndex(index int) {
	it.index = index
}
