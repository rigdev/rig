package capsule

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	container_name "github.com/google/go-containerregistry/pkg/name"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

type imageInfo struct {
	tag     string
	created time.Time
}

func CapsuleDeploy(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, rc rig.Client) error {
	dc, err := getDockerClient()
	if err != nil {
		return err
	}

	buildID, err = getBuildID(ctx, capsuleID, rc, dc)
	if err != nil {
		return err
	}
	res, err := rc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsuleID,
			Changes: []*capsule.Change{{
				Field: &capsule.Change_BuildId{BuildId: buildID},
			}},
		},
	})
	if err != nil {
		return err
	}
	cmd.Printf("Deploying build %v in rollout %v \n", buildID, res.Msg.GetRolloutId())
	return listenForEvents(ctx, res.Msg.GetRolloutId(), rc, capsuleID, cmd)
}

func getBuildID(ctx context.Context, capsuleID string, rc rig.Client, dc *client.Client) (string, error) {
	if buildID != "" && image != "" {
		return "", errors.New("not both --build-id and --image can be given")
	}

	if buildID != "" {
		// TODO Figure out pagination
		resp, err := rc.Capsule().ListBuilds(ctx, connect.NewRequest(&capsule.ListBuildsRequest{
			CapsuleId: capsuleID,
			Pagination: &model.Pagination{
				Offset:     0,
				Limit:      0,
				Descending: false,
			},
		}))
		if err != nil {
			return "", err
		}
		builds := resp.Msg.GetBuilds()
		return expandBuildID(ctx, builds, buildID)
	}

	if image != "" {
		return createBuild(ctx, rc, capsuleID, dc, image)
	}

	return promptForImageOrBuild(ctx, capsuleID, rc, dc)
}

func expandBuildID(ctx context.Context, builds []*capsule.Build, buildID string) (string, error) {
	if strings.HasPrefix(buildID, "sha256:") {
		return expandByDigestPrefix(buildID, builds)
	}
	if isHexString(buildID) {
		return expandByDigestPrefix("sha256:"+buildID, builds)
	}
	if strings.Contains(buildID, "@") {
		return expandByDigestName(buildID, builds)
	}
	if ref, err := container_name.NewTag(buildID); err == nil {
		return expandByLatestTag(ref, builds)
	}

	return "", errors.New("unable to parse buildID")
}

func expandByDigestName(buildID string, builds []*capsule.Build) (string, error) {
	idx := strings.Index(buildID, "@")
	name := buildID[:idx]
	digest := buildID[idx+1:]
	tag, err := container_name.NewTag(name)
	if err != nil {
		return "", err
	}
	var validBuilds []*capsule.Build
	for _, b := range builds {
		repoMatch := b.GetRepository() == tag.RepositoryStr()
		tagMatch := b.GetTag() == tag.TagStr()
		digMatch := strings.HasPrefix(b.GetDigest(), digest)
		if repoMatch && tagMatch && digMatch {
			validBuilds = append(validBuilds, b)
		}
	}

	if len(validBuilds) == 0 {
		return "", errors.New("no builds matched the image name and digest prefix")
	}
	if len(validBuilds) > 1 {
		return "", errors.New("the image name and digest prefix was not unique")
	}

	return validBuilds[0].GetBuildId(), nil
}

func expandByLatestTag(ref container_name.Reference, builds []*capsule.Build) (string, error) {
	var latest *capsule.Build
	for _, b := range builds {
		if b.GetRepository() != ref.Context().RepositoryStr() || b.GetTag() != ref.Identifier() {
			continue
		}
		if latest == nil || latest.CreatedAt.AsTime().Before(b.CreatedAt.AsTime()) {
			latest = b
		}
	}

	if latest == nil {
		return "", errors.New("no builds matched the given image name")
	}

	return latest.GetBuildId(), nil
}

func expandByDigestPrefix(digestPrefix string, builds []*capsule.Build) (string, error) {
	var validBuilds []*capsule.Build
	for _, b := range builds {
		if strings.HasPrefix(b.GetDigest(), digestPrefix) {
			validBuilds = append(validBuilds, b)
		}
	}
	if len(validBuilds) > 1 {
		return "", errors.New("digest prefix was not unique")
	}
	if len(validBuilds) == 0 {
		return "", errors.New("no builds had a matching digest prefix")
	}
	return validBuilds[0].GetBuildId(), nil
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

func createBuild(ctx context.Context, rc rig.Client, capsuleID string, dc *client.Client, image string) (string, error) {
	if strings.Contains(image, "@") {
		return "", errors.UnimplementedErrorf("referencing images by digest is not yet supported")
	}

	isLocalImage, _, err := utils.ImageExistsNatively(ctx, dc, image)
	if err != nil {
		return "", err
	}

	var digest string
	if isLocalImage {
		image, digest, err = pushLocalImageToDevRegistry(ctx, image, rc, dc)
		if err != nil {
			return "", err
		}
	}

	res, err := rc.Capsule().CreateBuild(ctx, &connect.Request[capsule.CreateBuildRequest]{
		Msg: &capsule.CreateBuildRequest{
			CapsuleId: capsuleID,
			Image:     image,
			Digest:    digest,
		},
	})
	if err != nil {
		return "", err
	}

	if res.Msg.GetCreatedNewBuild() {
		fmt.Println("created new build:", res.Msg.GetBuildId())
	} else {
		fmt.Println("build already exists, deploying existing build")
	}

	isLocalImage, _, err = utils.ImageExistsNatively(ctx, dc, res.Msg.BuildId)
	if err != nil {
		return "", err
	}

	return res.Msg.GetBuildId(), nil
}

func promptForImageOrBuild(ctx context.Context, capsuleID string, rc rig.Client, dc *client.Client) (string, error) {
	i, _, err := common.PromptSelect("Deploy from docker image or existing rig build?", []string{"Image", "Build"})
	if err != nil {
		return "", err
	}
	switch i {
	case 0:
		image, err := promptForImage(ctx, dc)
		if err != nil {
			return "", err
		}
		return createBuild(ctx, rc, capsuleID, dc, image)
	case 1:
		return promptForExistingBuild(ctx, capsuleID, rc)
	default:
		return "", errors.New("something went wrong")
	}
}

func promptForImage(ctx context.Context, dc *client.Client) (string, error) {
	ok, err := common.PromptConfirm("Deploy a local image?", true)
	if err != nil {
		return "", err
	}

	if ok {
		img, err := getDaemonImage(ctx, dc)
		if err != nil {
			return "", err
		}
		return img.tag, nil
	}

	image, err = common.PromptInput("Enter image:", common.ValidateImageOpt)
	if err != nil {
		return "", err
	}
	return image, nil
}

func getDaemonImage(ctx context.Context, dc *client.Client) (*imageInfo, error) {
	images, prompts, err := getImagePrompts(ctx, dc, "")
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, errors.New("no local docker images found")
	}
	idx, _, err := common.PromptSelect("Select image:", prompts, common.SelectEnableFilterOpt)
	if err != nil {
		return nil, err
	}
	return &images[idx], nil
}

func getImagePrompts(ctx context.Context, dc *client.Client, filter string) ([]imageInfo, []string, error) {
	res, err := dc.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.Arg("dangling", "false")),
	})
	if err != nil {
		return nil, nil, err
	}

	var images []imageInfo
	var prompts []string

	for _, image := range res {
		for _, tag := range image.RepoTags {
			t := time.Unix(image.Created, 0)
			ii, _, err := dc.ImageInspectWithRaw(ctx, tag)
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

	slices.SortFunc(images, func(i, j imageInfo) bool {
		return i.created.After(j.created)
	})

	for idx, image := range images {
		if idx >= 50 {
			break
		}
		t := time.Since(image.created).Round(time.Second)
		var timeString string
		if t.Hours() > 24 {
			days := int(t.Hours() / 24)
			timeString = fmt.Sprintf("%vd", days)
			t = t - time.Duration(days*24)*time.Hour
		}
		timeString = fmt.Sprintf("%s%v", timeString, t)

		prompts = append(prompts, fmt.Sprintf("%s [age: %v]", image.tag, timeString))
	}
	return images, prompts, nil
}

func promptForExistingBuild(ctx context.Context, capsuleID string, rc rig.Client) (string, error) {
	resp, err := rc.Capsule().ListBuilds(ctx, connect.NewRequest(&capsule.ListBuildsRequest{
		CapsuleId:  capsuleID,
		Pagination: &model.Pagination{},
	}))
	if err != nil {
		return "", err
	}
	builds := resp.Msg.GetBuilds()

	var rows [][]string
	for _, b := range builds {
		rows = append(rows, []string{
			fmt.Sprint(b.GetRepository(), ":", b.GetTag()),
			truncatedFixed(b.GetDigest(), 19),
			fmt.Sprint(time.Since(b.GetCreatedAt().AsTime()).Truncate(time.Second)),
		})
	}

	idx, err := common.PromptTableSelect(
		"Select a Rig build",
		rows,
		[]string{"Image name", "Digest", "Age"},
		common.SelectFuzzyFilterOpt,
	)
	if err != nil {
		return "", err
	}

	return builds[idx].GetBuildId(), nil
}

func listenForEvents(ctx context.Context, rolloutID uint64, rc rig.Client, capsuleID string, cmd *cobra.Command) error {
	eventCount := 0
	for {
		res, err := rc.Capsule().GetRollout(ctx, &connect.Request[capsule.GetRolloutRequest]{
			Msg: &capsule.GetRolloutRequest{
				CapsuleId: capsuleID,
				RolloutId: rolloutID,
			},
		})
		if err != nil {
			return err
		}

		eventRes, err := rc.Capsule().ListEvents(ctx, &connect.Request[capsule.ListEventsRequest]{
			Msg: &capsule.ListEventsRequest{
				CapsuleId: capsuleID,
				RolloutId: rolloutID,
				Pagination: &model.Pagination{
					Offset: uint32(eventCount),
				},
			},
		})
		if err != nil {
			return err
		}
		for _, event := range eventRes.Msg.GetEvents() {
			cmd.Printf("[%v] %v\n", event.GetCreatedAt().AsTime().Format(time.RFC822), event.GetMessage())
		}
		eventCount += len(eventRes.Msg.GetEvents())

		switch res.Msg.GetRollout().GetStatus().GetState() {
		case capsule.RolloutState_ROLLOUT_STATE_DONE:
			cmd.Println("Deployment complete")
			return nil
		case capsule.RolloutState_ROLLOUT_STATE_FAILED:
			cmd.Println("Deployment failed")
			return nil
		case capsule.RolloutState_ROLLOUT_STATE_ABORTED:
			cmd.Println("Deployment aborted")
			return nil
		}

		if len(eventRes.Msg.GetEvents()) == 0 {
			cmd.Println("Deploying build...")
		}

		time.Sleep(1 * time.Second)
	}
}

// TODO Should be supplied by FX instead
// TODO Currently we can't read from protected repositories as we don't properly read the credentials which the local docker CLI uses
func getDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
}

func pushLocalImageToDevRegistry(ctx context.Context, image string, client rig.Client, dc *client.Client) (string, string, error) {
	resp, err := client.Cluster().GetConfig(ctx, connect.NewRequest(&cluster.GetConfigRequest{}))
	if err != nil {
		return "", "", err
	}
	config := resp.Msg

	switch config.GetDevRegistry().(type) {
	case *cluster.GetConfigResponse_Docker:
		return image, "", nil
	}
	devRegistry := config.GetRegistry()
	if devRegistry == nil {
		return "", "", fmt.Errorf("no dev-registry configured.") // TODO Help the user with fixing this
	}

	newImageName, err := makeDevRegistryImageName(image, devRegistry.Host)
	if err != nil {
		return "", "", err
	}

	fmt.Printf("Pushing the image to the dev docker registry under the new name %q\n", newImageName)

	if err := dc.ImageTag(ctx, image, newImageName); err != nil {
		return "", "", err
	}

	digest, err := pushToDevRegistry(ctx, dc, newImageName, devRegistry.Host)
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

func pushToDevRegistry(ctx context.Context, dc *client.Client, image string, host string) (string, error) {
	ac := registry.AuthConfig{
		ServerAddress: host,
	}
	secret, err := json.Marshal(ac)
	if err != nil {
		return "", err
	}

	rc, err := dc.ImagePush(ctx, image, types.ImagePushOptions{
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
