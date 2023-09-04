package capsule

import (
	"context"

	"github.com/docker/distribution/reference"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) CreateBuild(ctx context.Context, capsuleID uuid.UUID, image, digest string, origin *capsule.Origin, labels map[string]string) (string, error) {
	if image == "" {
		return "", errors.InvalidArgumentErrorf("missing image")
	}

	ref, err := reference.ParseDockerRef(image)
	if err != nil {
		return "", errors.InvalidArgumentErrorf("%v", err)
	}

	tagged, ok := reference.TagNameOnly(ref).(reference.NamedTagged)
	if !ok {
		return "", errors.InvalidArgumentErrorf("invalid image tag")
	}

	if _, err := s.GetCapsule(ctx, capsuleID); err != nil {
		return "", err
	}

	by, err := s.as.GetAuthor(ctx)
	if err != nil {
		return "", err
	}

	b := &capsule.Build{
		BuildId:    tagged.String(),
		Digest:     digest,
		Repository: tagged.Name(),
		Tag:        tagged.Tag(),
		CreatedBy:  by,
		CreatedAt:  timestamppb.Now(),
		Origin:     origin,
		Labels:     labels,
	}

	if err := s.cr.CreateBuild(ctx, capsuleID, b); err != nil {
		return "", err
	}

	return tagged.String(), nil
}

func (s *Service) ListBuilds(ctx context.Context, capsuleID uuid.UUID, pagination *model.Pagination) (iterator.Iterator[*capsule.Build], uint64, error) {
	return s.cr.ListBuilds(ctx, pagination, capsuleID)
}
