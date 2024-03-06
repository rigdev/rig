package imagedeploy

import (
	"testing"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_expandBuildID(t *testing.T) {
	t1 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2000, 1, 1, 1, 0, 0, 0, time.UTC)
	builds := []*capsule.Image{
		{
			ImageId:    "registry.io/name:tag@sha256:0123456789",
			Digest:     "sha256:0123456789",
			Repository: "registry.io/name",
			Tag:        "tag",
			CreatedAt:  timestamppb.New(t1),
		},
		{
			ImageId:    "registry.io/name:tag@sha256:01234abcd",
			Digest:     "sha256:01234abcd",
			Repository: "registry.io/name",
			Tag:        "tag",
			CreatedAt:  timestamppb.New(t2),
		},
	}

	tests := []struct {
		name    string
		imageID string
		err     error
		res     string
	}{
		{
			name:    "exact match",
			imageID: "registry.io/name:tag@sha256:0123456789",
			err:     nil,
			res:     "registry.io/name:tag@sha256:0123456789",
		},
		{
			name:    "sha prefix",
			imageID: "sha256:01234567",
			err:     nil,
			res:     "registry.io/name:tag@sha256:0123456789",
		},
		{
			name:    "hex prefix",
			imageID: "01234567",
			err:     nil,
			res:     "registry.io/name:tag@sha256:0123456789",
		},
		{
			name:    "not unique prefix",
			imageID: "01234",
			err:     errors.New("digest prefix was not unique"),
			res:     "",
		},
		{
			name:    "no matching prefix",
			imageID: "012345f",
			err:     errors.New("no builds had a matching digest prefix"),
			res:     "",
		},
		{
			name:    "get latest by tag",
			imageID: "registry.io/name:tag",
			err:     nil,
			res:     "registry.io/name:tag@sha256:01234abcd",
		},
		{
			name:    "no build with tag",
			imageID: "registry.io/name:tag2",
			err:     errors.New("no builds matched the given image name"),
			res:     "",
		},
		{
			name:    "image name + digest prefix",
			imageID: "registry.io/name:tag@sha256:0123456",
			err:     nil,
			res:     "registry.io/name:tag@sha256:0123456789",
		},
		{
			name:    "malformed",
			imageID: "__+",
			err:     errors.New("unable to parse imageID"),
			res:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := expandBuildID(builds, tt.imageID)
			utils.ErrorEqual(t, tt.err, err)
			assert.Equal(t, tt.res, res)
		})
	}
}
