package utils

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func ImageExistsNatively(ctx context.Context, dc *client.Client, image string) (bool, string, error) {
	image = strings.TrimPrefix(image, "docker.io/library/")
	image = strings.TrimPrefix(image, "index.docker.io/library/")
	is, err := dc.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: image,
		}),
	})
	if err != nil {
		return false, "", err
	}

	if len(is) == 0 {
		return false, "", nil
	}

	return true, is[0].ID, nil
}
