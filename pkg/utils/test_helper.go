package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ErrorEqual(t *testing.T, expected error, err error) {
	if expected == nil {
		assert.Nil(t, err)
		return
	}
	assert.NotNil(t, err)
	assert.Equal(t, expected.Error(), err.Error())
}
