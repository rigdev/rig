package image

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/docker/docker/api/types/filters"
	docker_image "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	container_name "github.com/google/go-containerregistry/pkg/name"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/rigdev/rig-go-api/api/v1/image"
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
		Image:        imageID,
		IsKnownLocal: nil,
	}
	if remote {
		imageRef.IsKnownLocal = ptr.New(false)
	}
	return imageRef
}

func (c *Cmd) promptForImage(ctx context.Context) (imageRef, error) {
	var empty imageRef

	ok, err := c.Prompter.Confirm("Use a local image?", true)
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

	imageName, err := c.Prompter.Input("Enter image:", common.ValidateImageOpt)
	if err != nil {
		return empty, nil
	}
	return imageRef{
		Image:        imageName,
		IsKnownLocal: ptr.New(false),
	}, nil
}

func (c *Cmd) getDaemonImage(ctx context.Context) (*imageInfo, error) {
	images, prompts, err := c.getImagePrompts(ctx)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, errors.New("no local docker images found")
	}
	idx, err := c.Prompter.TableSelect(
		"Select image:", prompts, []string{"Image name", "Age"}, common.SelectEnableFilterOpt,
	)
	if err != nil {
		return nil, err
	}
	return &images[idx], nil
}

func (c *Cmd) getImagePrompts(ctx context.Context) ([]imageInfo, [][]string, error) {
	res, err := c.DockerClient.ImageList(ctx, docker_image.ListOptions{
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
			return 1
		}
		return -1
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

func (c *Cmd) createImageInner(ctx context.Context, capsuleID string, imageRef imageRef) (string, error) {
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

	res, err := c.Rig.Image().Add(ctx, connect.NewRequest(&image.AddRequest{
		CapsuleId:      capsuleID,
		Image:          imageRef.Image,
		Digest:         digest,
		SkipImageCheck: skipImageCheck,
		ProjectId:      c.Scope.GetCurrentContext().GetProject(),
	}),
	)
	if err != nil {
		return "", err
	}

	if res.Msg.GetAddedNewImage() {
		fmt.Println("Added new image:", res.Msg.GetImageId())
	} else {
		fmt.Println("Image already added, using existing image reference.")
	}

	return res.Msg.GetImageId(), nil
}

func (c *Cmd) pushLocalImageToDevRegistry(ctx context.Context, image string) (string, string, error) {
	resp, err := c.Rig.Cluster().GetConfigs(ctx, connect.NewRequest(&cluster.GetConfigsRequest{}))
	if err != nil {
		return "", "", err
	}

	clusters := resp.Msg.Clusters
	if len(clusters) != 1 {
		return "", "", errors.New("cannot push local images to dev registry if there are more than one cluster")
	}
	config := clusters[0]

	switch config.GetDevRegistry().(type) {
	case *cluster.GetConfigResponse_Docker:
		return image, "", nil
	}
	devRegistry := config.GetRegistry()
	if devRegistry == nil {
		return "", "", fmt.Errorf("no dev-registry configured") // TODO Help the user with fixing this
	}

	newImageName, err := makeDevRegistryImageName(image, devRegistry.Host)
	if err != nil {
		return "", "", err
	}

	fmt.Printf("Pushing the image to the dev docker registry under the new name %q\n", newImageName)

	if err := c.DockerClient.ImageTag(ctx, image, newImageName); err != nil {
		return "", "", err
	}

	digest, err := c.pushToDevRegistry(ctx, newImageName, devRegistry.Host)
	if err != nil {
		return "", "", err
	}

	return newImageName, digest, nil
}

func makeDevRegistryImageName(image string, devRegistryHost string) (string, error) {
	r, err := container_name.NewRegistry(devRegistryHost, container_name.Insecure)
	if err != nil {
		return "", err
	}
	ref, err := container_name.ParseReference(image)
	if err != nil {
		return "", err
	}
	repo := r.Repo(ref.Context().RepositoryStr())
	tag := repo.Tag(ref.Identifier())
	return tag.String(), nil
}

func (c *Cmd) pushToDevRegistry(ctx context.Context, img string, host string) (string, error) {
	ac := registry.AuthConfig{
		ServerAddress: host,
	}
	secret, err := json.Marshal(ac)
	if err != nil {
		return "", err
	}

	rc, err := c.DockerClient.ImagePush(ctx, img, docker_image.PushOptions{
		RegistryAuth: base64.StdEncoding.EncodeToString(secret),
	})
	if err != nil {
		return "", err
	}

	defer rc.Close()

	decoder := json.NewDecoder(rc)
	progressWriter := progress.NewWriter()
	progressWriter.SetAutoStop(true)
	trackers := map[string]*progress.Tracker{}

	go progressWriter.Render()
	var digest string
	for decoder.More() {
		var p dockerProgress
		if err := decoder.Decode(&p); err != nil {
			return "", err
		}
		if p.ID == "" || p.ProgressDetail.Total == 0 {
			continue
		}
		tracker, ok := trackers[p.ID]
		if !ok {
			tracker = &progress.Tracker{
				Message: p.ID,
				Total:   int64(p.ProgressDetail.Total),
				Units:   progress.UnitsBytes,
			}
			trackers[p.ID] = tracker
			progressWriter.AppendTracker(tracker)
		}
		if p.ProgressDetail.Current != 0 {
			tracker.SetValue(int64(p.ProgressDetail.Current))
		}
		if p.Aux.Digest != "" {
			digest = p.Aux.Digest
		}
	}

	return digest, nil
}

type dockerProgress struct {
	Status         string
	ID             string
	ProgressDetail struct {
		Current uint64
		Total   uint64
	}
	Aux struct {
		Tag    string
		Digest string
	}
}
