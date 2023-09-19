package capsule

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

type imageInfo struct {
	tag     string
	created time.Time
}

func CapsuleDeploy(ctx context.Context, cmd *cobra.Command, capsuleID CapsuleID, args []string, rc rig.Client) error {
	var err error
	if buildID == "" {
		dc, err := getDockerClient()
		if err != nil {
			return err
		}
		buildID, err = createBuild(ctx, rc, capsuleID, dc)
		if err != nil {
			return err
		}
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
			cmd.Printf("[%v] %v\n", event.GetCreatedAt().AsTime().Format(base.RFC3339MilliFixed), event.GetMessage())
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
func getDockerClient() (*client.Client, error) {
	var opts []client.Opt
	opts = append(opts, client.WithHostFromEnv())
	opts = append(opts, client.WithAPIVersionNegotiation())
	return client.NewClientWithOpts(opts...)
}

func getDaemonImage(ctx context.Context, dc *client.Client) (*imageInfo, error) {
	images, prompts, err := getImagePrompts(ctx, dc, "")
	if err != nil {
		return nil, err
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

func createBuild(ctx context.Context, rc rig.Client, capsuleID string, dc *client.Client) (string, error) {
	image, digest, err := getImageAndDigest(ctx, rc, dc)
	if err != nil {
		return "", err
	}
	res, err := rc.Capsule().CreateBuild(ctx, &connect.Request[capsule.CreateBuildRequest]{
		Msg: &capsule.CreateBuildRequest{
			CapsuleId: capsuleID,
			Image:     image,
			Digest:    digest,
		},
	})
	if errors.IsAlreadyExists(err) {
		fmt.Println("build already exists, deploying existing build")
		imgRef, err := container_name.ParseReference(image)
		if err != nil {
			return "", err
		}
		return imgRef.Name(), nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to create build: %q", err)
	}

	fmt.Println("Created build: ", buildID)
	return res.Msg.GetBuildId(), nil
}

func getImageAndDigest(ctx context.Context, rigClient rig.Client, dc *client.Client) (string, string, error) {
	if image != "" {
		return image, "", nil
	}

	isLocalImage := false
	ok, err := common.PromptConfirm("Deploy a local image?", true)
	if err != nil {
		return "", "", err
	}
	if ok {
		img, err := getDaemonImage(ctx, dc)
		if err != nil {
			return "", "", err
		}
		image = img.tag
		isLocalImage = true
	} else {
		image, err = common.PromptInput("Enter image:", common.ValidateImageOpt)
		if err != nil {
			return "", "", err
		}
		isLocalImage, _, err = utils.ImageExistsNatively(ctx, dc, image)
		if err != nil {
			return "", "", err
		}
	}

	if isLocalImage {
		return pushLocalImageToDevRegistry(ctx, image, rigClient, dc)
	}

	return image, "", nil
}

func pushLocalImageToDevRegistry(ctx context.Context, image string, client rig.Client, dc *client.Client) (string, string, error) {
	resp, err := client.Cluster().GetConfig(ctx, connect.NewRequest(&cluster.GetConfigRequest{}))
	if err != nil {
		return "", "", err
	}
	config := resp.Msg

	switch config.GetDevRegistry().(type) {
	case *cluster.GetConfigResponse_Docker:
		return "", "", nil
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
