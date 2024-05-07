package deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_containerNameFromEventFieldPath(t *testing.T) {
	assert.Equal(t, containerNameFromEventFieldPath("spec.initContainers{somename}"), "somename")
	assert.Equal(t, containerNameFromEventFieldPath("spec.initContainers{}"), "")
	assert.Equal(t, containerNameFromEventFieldPath("spec.initContainers{init-container}"), "init-container")
	assert.Equal(t, containerNameFromEventFieldPath("spec.containers{somename}"), "somename")
	assert.Equal(t, containerNameFromEventFieldPath("spec.containers{}"), "")
	assert.Equal(t, containerNameFromEventFieldPath("spec.containers{main-container}"), "main-container")
	assert.Equal(t, containerNameFromEventFieldPath("null"), "")
	assert.Equal(t, containerNameFromEventFieldPath(""), "")
}
