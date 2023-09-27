package base

import "github.com/spf13/cobra"

const (
	RFC3339NanoFixed  = "2006-01-02T15:04:05.000000000Z07:00"
	RFC3339MilliFixed = "2006-01-02T15:04:05.000Z07:00"
)

type Flag[T any] struct {
	Value T
	Name  string
	isSet bool
}

func (f Flag[T]) IsSet(cmd *cobra.Command) bool {
	return f.isSet || cmd.Flags().Lookup(f.Name).Changed
}

func (f Flag[T]) Set(value T) {
	f.isSet = true
	f.Value = value
}
