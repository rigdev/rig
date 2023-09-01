package iterator

import "io"

func Collect[T interface{}](it Iterator[T]) ([]T, error) {
	defer it.Close()

	var ts []T
	for {
		t, err := it.Next()
		if err == io.EOF {
			return ts, nil
		} else if err != nil {
			return nil, err
		}

		ts = append(ts, t)
	}
}
