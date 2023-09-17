package capsule

import (
	"context"
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	container_name "github.com/google/go-containerregistry/pkg/name"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

type imageInfo struct {
	tag     string
	created time.Time
}

func CapsuleDeploy(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, rc rig.Client) error {
	var err error
	if buildID == "" {
		buildID, err = createBuild(ctx, rc, capsuleID.String())
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

func listenForEvents(ctx context.Context, rolloutID uint64, rc rig.Client, capsuleID uuid.UUID, cmd *cobra.Command) error {
	eventCount := 0
	for {
		res, err := rc.Capsule().GetRollout(ctx, &connect.Request[capsule.GetRolloutRequest]{
			Msg: &capsule.GetRolloutRequest{
				CapsuleId: capsuleID.String(),
				RolloutId: rolloutID,
			},
		})
		if err != nil {
			return err
		}

		eventRes, err := rc.Capsule().ListEvents(ctx, &connect.Request[capsule.ListEventsRequest]{
			Msg: &capsule.ListEventsRequest{
				CapsuleId: capsuleID.String(),
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

func getDaemonImage(ctx context.Context) (*imageInfo, error) {
	var opts []client.Opt
	opts = append(opts, client.WithHostFromEnv())
	opts = append(opts, client.WithAPIVersionNegotiation())
	dc, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

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

	c := 0
	for _, image := range images {
		if c >= 50 {
			break
		}
		prompts = append(prompts, fmt.Sprintf("%s [age: %v]", image.tag, time.Since(image.created).Round(time.Second)))
		c++
	}
	return images, prompts, nil
}

func createBuild(ctx context.Context, rc rig.Client, capsuleID string) (string, error) {
	if image == "" {
		ok, err := common.PromptConfirm("Deploy a local image?", true)
		if err != nil {
			return "", err
		}
		if ok {
			img, err := getDaemonImage(ctx)
			if err != nil {
				return "", err
			}
			image = img.tag
		} else {
			image, err = common.PromptInput("Enter image:", common.ValidateImageOpt)
			if err != nil {
				return "", err
			}
		}
	}

	res, err := rc.Capsule().CreateBuild(ctx, &connect.Request[capsule.CreateBuildRequest]{
		Msg: &capsule.CreateBuildRequest{
			CapsuleId: capsuleID,
			Image:     image,
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
		return "", err
	}

	fmt.Println("Created build: ", buildID)
	return res.Msg.GetBuildId(), nil
}
