package builddeploy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	container_name "github.com/google/go-containerregistry/pkg/name"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/capsule/rollout"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) deploy(ctx context.Context, cmd *cobra.Command, _ []string) error {
	buildID, err := c.getBuildID(ctx, capsule_cmd.CapsuleID)
	if err != nil {
		return err
	}

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{{
				Field: &capsule.Change_BuildId{BuildId: buildID},
			}},
			ProjectId:     c.Cfg.GetProject(),
			EnvironmentId: base.Flags.Environment,
		},
	}

	res, err := c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			res, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, c.Cfg, capsule_cmd.CapsuleID, req)
		} else {
			res, err = capsule_cmd.PromptAbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, c.Cfg, req)
		}
	}
	if err != nil {
		return err
	}

	cmd.Printf("Deploying build %v in rollout %v \n", buildID, res.Msg.GetRolloutId())
	return c.listenForEvents(ctx, res.Msg.GetRolloutId(), capsule_cmd.CapsuleID)
}

func (c *Cmd) getBuildID(ctx context.Context, capsuleID string) (string, error) {
	if buildID != "" && image != "" {
		return "", errors.New("not both --build-id and --image can be given")
	}

	if buildID != "" {
		// TODO Figure out pagination
		resp, err := c.Rig.Capsule().ListBuilds(ctx, connect.NewRequest(&capsule.ListBuildsRequest{
			CapsuleId: capsuleID,
		}))
		if err != nil {
			return "", err
		}
		builds := resp.Msg.GetBuilds()
		return expandBuildID(builds, buildID)
	}

	if image != "" {
		return c.createBuildInner(ctx, capsuleID, imageRefFromFlags())
	}

	return c.promptForImageOrBuild(ctx, capsuleID)
}

func expandBuildID(builds []*capsule.Build, buildID string) (string, error) {
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
		repoMatch := b.GetRepository() == fmt.Sprintf("%s/%s", tag.RegistryStr(), tag.RepositoryStr())
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
		if b.GetRepository() != fmt.Sprintf("%s/%s", ref.Context().RegistryStr(), ref.Context().RepositoryStr()) ||
			b.GetTag() != ref.Identifier() {
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

func (c *Cmd) promptForImageOrBuild(ctx context.Context, capsuleID string) (string, error) {
	i, _, err := common.PromptSelect("Deploy from docker image or existing rig build?", []string{"Image", "Build"})
	if err != nil {
		return "", err
	}
	switch i {
	case 0:
		imgRef, err := c.promptForImage(ctx)
		if err != nil {
			return "", err
		}
		return c.createBuildInner(ctx, capsuleID, imgRef)
	case 1:
		return c.promptForExistingBuild(ctx, capsuleID)
	default:
		return "", errors.New("something went wrong")
	}
}

func (c *Cmd) promptForExistingBuild(ctx context.Context, capsuleID string) (string, error) {
	resp, err := c.Rig.Capsule().ListBuilds(ctx, connect.NewRequest(&capsule.ListBuildsRequest{
		CapsuleId:  capsuleID,
		Pagination: &model.Pagination{},
	}))
	if err != nil {
		return "", err
	}
	builds := resp.Msg.GetBuilds()
	slices.SortFunc(builds, func(b1, b2 *capsule.Build) int {
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

	if len(builds) == 0 {
		return "", errors.New("capsule has no builds")
	}

	var rows [][]string
	for _, b := range builds {
		rows = append(rows, []string{
			fmt.Sprint(b.GetRepository(), ":", b.GetTag()),
			capsule_cmd.TruncatedFixed(b.GetDigest(), 19),
			common.FormatDuration(time.Since(b.GetCreatedAt().AsTime())),
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

func (c *Cmd) listenForEvents(ctx context.Context, rolloutID uint64, capsuleID string) error {
	eventCount := 0
	for {
		res, err := c.Rig.Capsule().GetRollout(ctx, &connect.Request[capsule.GetRolloutRequest]{
			Msg: &capsule.GetRolloutRequest{
				CapsuleId: capsuleID,
				RolloutId: rolloutID,
			},
		})
		if err != nil {
			return err
		}

		eventRes, err := c.Rig.Capsule().ListEvents(ctx, &connect.Request[capsule.ListEventsRequest]{
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
			fmt.Printf("[%v] %v\n", event.GetCreatedAt().AsTime().Format(base.RFC3339MilliFixed), event.GetMessage())
		}
		eventCount += len(eventRes.Msg.GetEvents())

		switch res.Msg.GetRollout().GetRolloutStatus().GetState() {
		case rollout.State_STATE_RUNNING, rollout.State_STATE_STOPPED:
			fmt.Printf("[%v] %v\n", time.Now().UTC().Format(time.RFC822), "Deployment complete")
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}

func (c *Cmd) pushLocalImageToDevRegistry(ctx context.Context, image string) (string, string, error) {
	resp, err := c.Rig.Cluster().GetConfig(ctx, connect.NewRequest(&cluster.GetConfigRequest{}))
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
