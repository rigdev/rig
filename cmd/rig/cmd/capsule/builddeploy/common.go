package builddeploy

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/utils"
)

type imageRef struct {
	Image string
	// &true: we know it's local
	// &false: we know it's remote
	// nil: we don't know
	IsKnownLocal *bool
}

func imageRefFromFlags() imageRef {
	imageRef := imageRef{
		Image:        image,
		IsKnownLocal: nil,
	}
	if remote {
		imageRef.IsKnownLocal = ptr.New(false)
	}
	return imageRef
}

func (c Cmd) promptForImage(ctx context.Context) (imageRef, error) {
	var empty imageRef

	ok, err := common.PromptConfirm("Use a local image?", true)
	if err != nil {
		return empty, err
	}

	if ok {
		img, err := c.getDaemonImage(ctx)
		if err != nil {
			return empty, err
		}
		return imageRef{
			Image:        img.tag,
			IsKnownLocal: ptr.New(true),
		}, nil
	}

	image, err = common.PromptInput("Enter image:", common.ValidateImageOpt)
	if err != nil {
		return empty, nil
	}
	return imageRef{
		Image:        image,
		IsKnownLocal: ptr.New(false),
	}, nil
}

func (c Cmd) getDaemonImage(ctx context.Context) (*imageInfo, error) {
	images, prompts, err := c.getImagePrompts(ctx, "")
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, errors.New("no local docker images found")
	}
	idx, err := common.PromptTableSelect("Select image:", prompts, []string{"Image name", "Age"}, common.SelectEnableFilterOpt)
	if err != nil {
		return nil, err
	}
	return &images[idx], nil
}

func (c Cmd) getImagePrompts(ctx context.Context, filter string) ([]imageInfo, [][]string, error) {
	res, err := c.DockerClient.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.Arg("dangling", "false")),
	})
	if err != nil {
		return nil, nil, err
	}

	var images []imageInfo
	var prompts [][]string

	for _, image := range res {
		for _, tag := range image.RepoTags {
			t := time.Unix(image.Created, 0)
			ii, _, err := c.DockerClient.ImageInspectWithRaw(ctx, tag)
			if err != nil {
				return nil, nil, err
			}
			if !ii.Metadata.LastTagTime.IsZero() {
				t = ii.Metadata.LastTagTime
			}
			images = append(images, imageInfo{
				tag:     tag,
				created: t,
			})
		}
	}

	slices.SortFunc(images, func(i, j imageInfo) int {
		if i.created.Equal(j.created) {
			return 0
		}
		if i.created.Before(j.created) {
			return -1
		}
		return 1
	})

	for idx, image := range images {
		if idx >= 50 {
			break
		}
		t := time.Since(image.created)
		prompts = append(prompts, []string{image.tag, common.FormatDuration(t)})
	}
	return images, prompts, nil
}

type imageInfo struct {
	tag     string
	created time.Time
}

func (c Cmd) createBuildInner(ctx context.Context, capsuleID string, imageRef imageRef) (string, error) {
	if strings.Contains(imageRef.Image, "@") {
		return "", errors.UnimplementedErrorf("referencing images by digest is not yet supported")
	}

	var err error
	var isLocalImage bool
	if imageRef.IsKnownLocal == nil {
		isLocalImage, _, err = utils.ImageExistsNatively(ctx, c.DockerClient, imageRef.Image)
		if err != nil {
			return "", err
		}
	} else {
		isLocalImage = *imageRef.IsKnownLocal
	}

	var digest string
	if isLocalImage {
		imageRef.Image, digest, err = c.pushLocalImageToDevRegistry(ctx, imageRef.Image)
		if err != nil {
			return "", err
		}
	}

	res, err := c.Rig.Capsule().CreateBuild(ctx, &connect.Request[capsule.CreateBuildRequest]{
		Msg: &capsule.CreateBuildRequest{
			CapsuleId: capsuleID,
			Image:     imageRef.Image,
			Digest:    digest,
		},
	})
	if err != nil {
		return "", err
	}

	if res.Msg.GetCreatedNewBuild() {
		fmt.Println("Created new build:", res.Msg.GetBuildId())
	} else {
		fmt.Println("Build already exists, using existing build")
	}

	return res.Msg.GetBuildId(), nil
}
