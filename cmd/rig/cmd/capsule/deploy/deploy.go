package deploy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/registry"
	container_name "github.com/google/go-containerregistry/pkg/name"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	api_image "github.com/rigdev/rig-go-api/api/v1/image"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/ptr"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

func parseEnvironmentSource(value string) (capsule.EnvironmentSource_Kind, string, error) {
	var kind capsule.EnvironmentSource_Kind
	parts := strings.SplitN(value, "/", 2)
	if len(parts) != 2 {
		return kind, "", errors.InvalidArgumentErrorf("invalid --env-source format: %s", value)
	}

	switch strings.ToLower(parts[0]) {
	case "configmap":
		kind = capsule.EnvironmentSource_KIND_CONFIG_MAP
	case "secret":
		kind = capsule.EnvironmentSource_KIND_SECRET
	default:
		return kind, "", errors.InvalidArgumentErrorf("invalid --env-source kind, must be ConfigMap or Secret: %s", value)
	}

	return kind, parts[1], nil
}

func (c *Cmd) deploy(ctx context.Context, cmd *cobra.Command, args []string) error {
	var changes []*capsule.Change

	// Annotations.
	for _, key := range removeAnnotations {
		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_RemoveAnnotation{
				RemoveAnnotation: key,
			},
		})
	}
	for key, value := range annotations {
		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_SetAnnotation{
				SetAnnotation: &capsule.Change_KeyValue{
					Name:  key,
					Value: value,
				},
			},
		})
	}

	// Environment variables.
	for _, key := range removeEnvironmentVariables {
		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_RemoveEnvironmentVariable{
				RemoveEnvironmentVariable: key,
			},
		})
	}
	for key, value := range environmentVariables {
		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_SetEnvironmentVariable{
				SetEnvironmentVariable: &capsule.Change_KeyValue{
					Name:  key,
					Value: value,
				},
			},
		})
	}

	// Environment sources.
	for _, value := range removeEnvironmentSources {
		kind, name, err := parseEnvironmentSource(value)
		if err != nil {
			return err
		}
		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_RemoveEnvironmentSource{
				RemoveEnvironmentSource: &capsule.EnvironmentSource{
					Kind: kind,
					Name: name,
				},
			},
		})
	}
	for _, value := range environmentSources {
		kind, name, err := parseEnvironmentSource(value)
		if err != nil {
			return err
		}
		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_SetEnvironmentSource{
				SetEnvironmentSource: &capsule.EnvironmentSource{
					Kind: kind,
					Name: name,
				},
			},
		})
	}

	// Image.
	if imageID != "" {
		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_AddImage_{
				AddImage: &capsule.Change_AddImage{
					Image: imageID,
				},
			},
		})
	}

	// Network interfaces.
	for _, file := range networkInterfaces {
		bs, err := os.ReadFile(file)
		if err != nil {
			return errors.InvalidArgumentErrorf("errors reading network interface: %v", err)
		}

		raw, err := yaml.YAMLToJSON(bs)
		if err != nil {
			return err
		}

		ci := &capsule.Interface{}
		if err := protojson.Unmarshal(raw, ci); err != nil {
			return err
		}

		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_SetInterface{
				SetInterface: ci,
			},
		})
	}

	for _, name := range removeNetworkInterfaces {
		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_RemoveInterface{
				RemoveInterface: name,
			},
		})
	}

	// Config Files
	for _, target := range removeConfigFiles {
		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_RemoveConfigFile{
				RemoveConfigFile: target,
			},
		})
	}

	for _, configFile := range configFiles {
		var target string
		var source string
		var secret bool
		for _, opt := range strings.Split(configFile, ",") {
			opt = strings.TrimSpace(opt)
			if v, ok := strings.CutPrefix(opt, "path="); ok {
				target = v
			} else if v, ok := strings.CutPrefix(opt, "src="); ok {
				source = v
			} else if opt == "secret" {
				secret = true
			} else {
				return errors.InvalidArgumentErrorf("invalid config-file argument: %v", configFile)
			}
		}

		if !path.IsAbs(target) {
			return errors.InvalidArgumentErrorf("config-file path is not absolute: %v", target)
		}

		if path.Clean(target) != target {
			return errors.InvalidArgumentErrorf("config-file path is not clean: %v should be %s", target, path.Clean(target))
		}

		if strings.HasSuffix(target, "/") {
			return errors.InvalidArgumentErrorf("config-file path should not end with a '/': %v", target)
		}

		bs, err := os.ReadFile(source)
		if err != nil {
			return err
		}

		if !utf8.Valid(bs) {
			return errors.InvalidArgumentErrorf("source file is not valid UTF-8: %v", source)
		}

		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_SetConfigFile{
				SetConfigFile: &capsule.Change_ConfigFile{
					Content:  bs,
					Path:     target,
					IsSecret: secret,
				},
			},
		})
	}

	// Replicas.
	if cmd.Flag("replicas").Changed {
		if replicas < 0 {
			return errors.InvalidArgumentErrorf("number of replicas cannot be negative: %v", replicas)
		}

		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_Replicas{
				Replicas: uint32(replicas),
			},
		})
	}

	// Command and arguments.
	if idx := cmd.ArgsLenAtDash(); idx >= 0 {
		extraArgs := args[idx:]
		args = args[:idx]

		if len(extraArgs) == 0 {
			// Clear the command.
			changes = append(changes, &capsule.Change{
				Field: &capsule.Change_CommandArguments_{
					CommandArguments: &capsule.Change_CommandArguments{},
				},
			})
		} else {
			changes = append(changes, &capsule.Change{
				Field: &capsule.Change_CommandArguments_{
					CommandArguments: &capsule.Change_CommandArguments{
						Command: extraArgs[0],
						Args:    extraArgs[1:],
					},
				},
			})
		}
	}

	if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
		return err
	}

	// Capsule name.
	capsuleName := ""
	if len(args) > 0 {
		capsuleName = args[0]
	}

	if len(capsuleName) == 0 {
		if !c.Scope.IsInteractive() {
			return errors.InvalidArgumentErrorf("missing capsule name argument")
		}

		name, err := capsule_cmd.SelectCapsule(ctx, c.Rig, c.Prompter, c.Scope)
		if err != nil {
			return err
		}

		capsuleName = name
	}

	if len(changes) == 0 {
		if !c.Scope.IsInteractive() {
			return errors.InvalidArgumentErrorf("no changes to deploy")
		}

		imageID, err := c.GetImageID(ctx, capsuleName)
		if err != nil {
			return err
		}

		changes = append(changes, &capsule.Change{
			Field: &capsule.Change_ImageId{ImageId: imageID},
		})
	}

	revision, rolloutID, err := capsule_cmd.Deploy(
		ctx,
		c.Rig,
		c.Scope,
		capsuleName,
		changes,
		true,
		forceOverride,
		currentRolloutID,
	)
	if err != nil {
		return err
	}

	cmd.Printf("Deploying to capsule %v\n", capsuleName)

	if noWait {
		return nil
	}

	return capsule_cmd.WaitForRollout(ctx, c.Rig, c.Scope, capsuleName, revision, rolloutID)
}

func (c *Cmd) GetImageID(ctx context.Context, capsuleID string) (string, error) {
	if imageID != "" {
		// TODO Figure out pagination
		resp, err := c.Rig.Image().List(ctx, connect.NewRequest(&api_image.ListRequest{
			CapsuleId: capsuleID,
			ProjectId: flags.GetProject(c.Scope),
		}))
		if err != nil {
			return "", err
		}
		images := resp.Msg.GetImages()
		return expandImageID(images, imageID)
	}

	return c.promptForDockerOrImage(ctx, capsuleID)
}

func expandImageID(images []*capsule.Image, imageID string) (string, error) {
	if strings.HasPrefix(imageID, "sha256:") {
		return expandByDigestPrefix(imageID, images)
	}
	if isHexString(imageID) {
		return expandByDigestPrefix("sha256:"+imageID, images)
	}
	if strings.Contains(imageID, "@") {
		return expandByDigestName(imageID, images)
	}
	if ref, err := container_name.NewTag(imageID); err == nil {
		return expandByLatestTag(ref, images)
	}

	return "", errors.New("unable to parse image")
}

func expandByDigestName(imageID string, images []*capsule.Image) (string, error) {
	idx := strings.Index(imageID, "@")
	name := imageID[:idx]
	digest := imageID[idx+1:]
	tag, err := container_name.NewTag(name)
	if err != nil {
		return "", err
	}
	var validImages []*capsule.Image
	for _, b := range images {
		repoMatch := b.GetRepository() == fmt.Sprintf("%s/%s", tag.RegistryStr(), tag.RepositoryStr())
		tagMatch := b.GetTag() == tag.TagStr()
		digMatch := strings.HasPrefix(b.GetDigest(), digest)
		if repoMatch && tagMatch && digMatch {
			validImages = append(validImages, b)
		}
	}

	if len(validImages) == 0 {
		return "", errors.New("no images matched the image name and digest prefix")
	}
	if len(validImages) > 1 {
		return "", errors.New("the image name and digest prefix was not unique")
	}

	return validImages[0].GetImageId(), nil
}

func expandByLatestTag(ref container_name.Reference, images []*capsule.Image) (string, error) {
	var latest *capsule.Image
	for _, i := range images {
		if i.GetRepository() != fmt.Sprintf("%s/%s", ref.Context().RegistryStr(), ref.Context().RepositoryStr()) ||
			i.GetTag() != ref.Identifier() {
			continue
		}
		if latest == nil || latest.CreatedAt.AsTime().Before(i.CreatedAt.AsTime()) {
			latest = i
		}
	}

	if latest == nil {
		return "", errors.New("no images matched the given image name")
	}

	return latest.GetImageId(), nil
}

func expandByDigestPrefix(digestPrefix string, images []*capsule.Image) (string, error) {
	var validImages []*capsule.Image
	for _, b := range images {
		if strings.HasPrefix(b.GetDigest(), digestPrefix) {
			validImages = append(validImages, b)
		}
	}
	if len(validImages) > 1 {
		return "", errors.New("digest prefix was not unique")
	}
	if len(validImages) == 0 {
		return "", errors.New("no images had a matching digest prefix")
	}
	return validImages[0].GetImageId(), nil
}

func isHexString(s string) bool {
	s = strings.ToLower(s)
	for _, c := range s {
		if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'f')) {
			return false
		}
	}
	return true
}

func (c *Cmd) promptForDockerOrImage(ctx context.Context, capsuleID string) (string, error) {
	i, _, err := c.Prompter.Select(
		"Deploy from docker image rig-registered image?",
		[]string{"Docker", "Rig registered"},
	)
	if err != nil {
		return "", err
	}
	switch i {
	case 0:
		imgRef, err := c.promptForImage(ctx)
		if err != nil {
			return "", err
		}
		return c.createImageInner(ctx, capsuleID, imgRef)
	case 1:
		return c.promptForExistingImage(ctx, capsuleID)
	default:
		return "", errors.New("something went wrong")
	}
}

func (c *Cmd) promptForExistingImage(ctx context.Context, capsuleID string) (string, error) {
	resp, err := c.Rig.Image().List(ctx, connect.NewRequest(&api_image.ListRequest{
		CapsuleId:  capsuleID,
		Pagination: &model.Pagination{},
		ProjectId:  flags.GetProject(c.Scope),
	}))
	if err != nil {
		return "", err
	}
	images := resp.Msg.GetImages()
	slices.SortFunc(images, func(b1, b2 *capsule.Image) int {
		t1 := b1.CreatedAt.AsTime()
		t2 := b2.CreatedAt.AsTime()
		if t1.Equal(t2) {
			return 0
		}
		if t1.Before(t2) {
			return 1
		}
		return -1
	})

	if len(images) == 0 {
		return "", errors.New("capsule has no images")
	}

	var rows [][]string
	for _, b := range images {
		rows = append(rows, []string{
			fmt.Sprint(b.GetRepository(), ":", b.GetTag()),
			capsule_cmd.TruncatedFixed(b.GetDigest(), 19),
			common.FormatDuration(time.Since(b.GetCreatedAt().AsTime())),
		})
	}

	idx, err := c.Prompter.TableSelect(
		"Select a Rig image",
		rows,
		[]string{"Image name", "Digest", "Age"},
		common.SelectFuzzyFilterOpt,
	)
	if err != nil {
		return "", err
	}

	return images[idx].GetImageId(), nil
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

func (c *Cmd) pushToDevRegistry(ctx context.Context, image string, host string) (string, error) {
	ac := registry.AuthConfig{
		ServerAddress: host,
	}
	secret, err := json.Marshal(ac)
	if err != nil {
		return "", err
	}

	rc, err := c.DockerClient.ImagePush(ctx, image, types.ImagePushOptions{
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

type imageRef struct {
	Image string
	// &true: we know it's local
	// &false: we know it's remote
	// nil: we don't know
	IsKnownLocal *bool
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

	res, err := c.Rig.Image().Add(ctx, connect.NewRequest(&api_image.AddRequest{
		CapsuleId:      capsuleID,
		Image:          imageRef.Image,
		Digest:         digest,
		SkipImageCheck: skipImageCheck,
		ProjectId:      flags.GetProject(c.Scope),
	}))
	if err != nil {
		return "", err
	}

	if res.Msg.GetAddedNewImage() {
		fmt.Println("Added new image:", res.Msg.GetImageId())
	} else {
		fmt.Println("Image already exists, using existing image")
	}

	return res.Msg.GetImageId(), nil
}
