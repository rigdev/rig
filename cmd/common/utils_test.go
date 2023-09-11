package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FormatIntToSI(t *testing.T) {
	tests := []struct {
		name string
		n    uint64
		d    int
		s    string
	}{
		{
			name: "< 1000",
			n:    234,
			d:    3,
			s:    "234",
		},
		{
			name: "k",
			n:    43000,
			d:    3,
			s:    "43k",
		},
		{
			name: "k, decimals",
			n:    12345,
			d:    3,
			s:    "12.3k",
		},
		{
			name: "M",
			n:    12_000_000,
			d:    3,
			s:    "12M",
		},
		{
			name: "M decimals",
			n:    12_345_678,
			d:    3,
			s:    "12.3M",
		},
		{
			name: "G",
			n:    123_000_000_000,
			d:    3,
			s:    "123G",
		},
		{
			name: "G round away decimals",
			n:    123_456_000_000,
			d:    3,
			s:    "123G",
		},
		{
			name: "T",
			n:    1_456_000_000_000,
			d:    3,
			s:    "1.46T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := FormatIntToSI(uint64(tt.n), tt.d)
			assert.Equal(t, tt.s, ss)
		})
	}
}

func Test_ToStringWithSignificantDigits(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		d    int
		s    string
	}{
		{
			name: "simple",
			f:    123,
			d:    3,
			s:    "123",
		},
		{
			name: "integer fewer digits",
			f:    12,
			d:    3,
			s:    "12",
		},
		{
			name: "fractional input not fractional output",
			f:    123.456,
			d:    3,
			s:    "123",
		},
		{
			name: "fractional input fractional output",
			f:    123.456,
			d:    4,
			s:    "123.5",
		},
		{
			name: "input less than 1",
			f:    0.123456,
			d:    3,
			s:    "0.123",
		},
		{
			name: "input less than 1 with zeros",
			f:    0.0123456,
			d:    3,
			s:    "0.012",
		},
		{
			name: "input larger than precision",
			f:    12345,
			d:    3,
			s:    "12300",
		},
		{
			name: "becomes zero",
			f:    0.000123,
			d:    3,
			s:    "0",
		},
		{
			name: "negative",
			f:    -0.123456,
			d:    3,
			s:    "-0.123",
		},
		{
			name: "negative integer",
			f:    -123456,
			d:    3,
			s:    "-123000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := ToStringWithSignificantDigits(tt.f, tt.d)
			assert.Equal(t, tt.s, ss)
		})
	}
}
