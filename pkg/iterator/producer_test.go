package iterator

import (
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Producer_Error(t *testing.T) {
	p := NewProducer[int]()

	go func() {
		for i := 0; i < 3; i++ {
			p.Value(i)
		}
		p.Error(io.EOF)
	}()

	for i := 0; i < 3; i++ {
		n, err := p.Next()
		require.NoError(t, err)
		require.Equal(t, i, n)
	}

	n, err := p.Next()
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, 0, n)
}

func Test_Producer_ValueAfterError(t *testing.T) {
	p := NewProducer[int]()

	go func() {
		p.Value(5)
		p.Error(io.EOF)
		p.Value(6)
	}()

	n, err := p.Next()
	require.NoError(t, err)
	require.Equal(t, 5, n)

	n, err = p.Next()
	require.EqualError(t, err, "EOF")
	require.Equal(t, 0, n)
}

func Test_Producer_ErrorAfterError(t *testing.T) {
	p := NewProducer[int]()

	go func() {
		p.Error(io.EOF)
		p.Error(io.ErrUnexpectedEOF)
	}()

	n, err := p.Next()
	require.EqualError(t, err, "EOF")
	require.Equal(t, 0, n)
}

func Test_Producer_CloseAbortsValue(t *testing.T) {
	p := NewProducer[int]()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		require.NoError(t, p.Value(5))
		require.EqualError(t, p.Value(6), "context canceled")
		wg.Done()
	}()

	n, err := p.Next()
	require.NoError(t, err)
	require.Equal(t, 5, n)

	p.Close()

	wg.Wait()
}

func Test_Producer_NilErrorIsEOF(t *testing.T) {
	p := NewProducer[int]()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		p.Error(nil)
		wg.Done()
	}()

	n, err := p.Next()
	require.EqualError(t, err, "EOF")
	require.Equal(t, 0, n)

	n, err = p.Next()
	require.EqualError(t, err, "EOF")
	require.Equal(t, 0, n)

	wg.Wait()
}
