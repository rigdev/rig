package utils

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/name"
)

func ImageExistsNatively(ctx context.Context, dc *client.Client, img string) (bool, string, error) {
	img = strings.TrimPrefix(img, "docker.io/library/")
	img = strings.TrimPrefix(img, "index.docker.io/library/")
	is, err := dc.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: img,
		}),
	})
	if err != nil {
		return false, "", err
	}

	if len(is) == 0 {
		return false, "", nil
	}

	ref, err := name.ParseReference(img)
	if err != nil {
		return false, "", err
	}

	// A local build which has never been pushed to a registry has no digest
	// See https://github.com/moby/moby/issues/16482#issuecomment-149285106
	// A remote build pulled to local will look like a local build (as it is returned by ImageList)
	// but will have a digest
	// This distinguishes between these two cases
	var digest string
	for _, refStrWithDigest := range is[0].RepoDigests {
		refWithDigest, err := name.ParseReference(refStrWithDigest)
		if err != nil {
			return false, "", err
		}
		if ref.Context().RepositoryStr() == refWithDigest.Context().RepositoryStr() {
			digest = refWithDigest.Identifier()
			break
		}
	}

	return true, digest, nil
}
