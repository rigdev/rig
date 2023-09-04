package capsule

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/distribution/distribution/v3/reference"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func CapsulePush(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, nc rig.Client, cfg *base.Config) error {
	var err error
	if image == "" {
		image, err = utils.PromptGetInput("Enter Image", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	ref, err := reference.ParseDockerRef(image)
	if err != nil {
		return err
	}

	// Generate a new tag for the build.
	tag := strings.ReplaceAll(uuid.New().String()[:24], "-", "")

	rigImage := fmt.Sprint("localhost:5001/rig/", reference.Path(ref), ":", tag)

	cmd.Printf("Pushing image '%s' as '%s\n", image, rigImage)

	if err := run(ctx, "docker", "image", "tag", image, rigImage); err != nil {
		return err
	}

	rigRef, err := reference.ParseNamed(rigImage)
	if err != nil {
		return err
	}

	pw := progress.NewWriter()
	pw.SetAutoStop(true)
	pt := &progress.Tracker{
		Message: "pushing image",
	}
	pw.AppendTracker(pt)

	go pw.Render()

	bs, err := output(ctx, "docker", "image", "push", rigImage)
	if err != nil {
		return err
	}

	digests := regexp.MustCompile("digest: (sha256:.*) size:").FindStringSubmatch(string(bs))
	var digest string
	if len(digests) > 1 {
		digest = digests[1]
	}

	pt.MarkAsDone()

	if _, err := output(ctx, "docker", "image", "rm", rigImage); err != nil {
		return err
	}

	buildID := tag

	if _, err := nc.Capsule().CreateBuild(ctx, &connect.Request[capsule.CreateBuildRequest]{
		Msg: &capsule.CreateBuildRequest{
			CapsuleId: capsuleID.String(),
			Image:     rigRef.Name(),
			Digest:    digest,
		},
	}); err != nil {
		return err
	}

	if deploy {
		if _, err := nc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
			Msg: &capsule.DeployRequest{
				CapsuleId: capsuleID.String(),
				Changes: []*capsule.Change{{
					Field: &capsule.Change_BuildId{
						BuildId: buildID,
					},
				}},
			},
		}); err != nil {
			return err
		}

		cmd.Printf("Deployed build %v\n", buildID)
	} else {
		cmd.Printf("Image available as build %v\n", buildID)
	}

	return nil
}

func output(ctx context.Context, name string, args ...string) ([]byte, error) {
	p := exec.CommandContext(ctx, name, args...)
	p.Stderr = os.Stderr
	return p.Output()
}

func run(ctx context.Context, name string, args ...string) error {
	p := exec.CommandContext(ctx, name, args...)
	p.Stdout = os.Stdout
	p.Stderr = os.Stderr
	return p.Run()
}
