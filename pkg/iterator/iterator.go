package iterator

// Iterator is a generic interface for streaming a series of item, with
// support for back-pressure.
type Iterator[T interface{}] interface {
	// Next returns the next item, or an error. All errors are terminal; no more values
	// will arrive after the first error.
	Next() (T, error)

	// Close the consumer side of the Iterator. Must always be called when
	// the iterator is no longer used, to release resources.
	Close()
}
