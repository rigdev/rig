package capsule

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) CreateBuild(ctx context.Context, capsuleID uuid.UUID, image, digest string, origin *capsule.Origin, labels map[string]string, validateImage bool) (string, error) {
	if image == "" {
		return "", errors.InvalidArgumentErrorf("missing image")
	}

	ref, err := name.ParseReference(image)
	if err != nil {
		return "", errors.InvalidArgumentErrorf("%v", err)
	}

	if validateImage {
		d, err := s.validateImage(ctx, ref)
		if err != nil {
			return "", err
		}

		if digest != "" && digest != d {
			return "", errors.InvalidArgumentErrorf("provided digest doesn't match image")
		}

		digest = d
		if ref.Identifier() == "latest" {
			ref, err = name.ParseReference(ref.Context().String() + "@" + digest)
			if err != nil {
				return "", err
			}
		}
	}

	if _, err := s.GetCapsule(ctx, capsuleID); err != nil {
		return "", err
	}

	by, err := s.as.GetAuthor(ctx)
	if err != nil {
		return "", err
	}

	b := &capsule.Build{
		BuildId:    ref.Name(),
		Digest:     digest,
		Repository: ref.Context().RepositoryStr(),
		Tag:        ref.Identifier(),
		CreatedBy:  by,
		CreatedAt:  timestamppb.Now(),
		Origin:     origin,
		Labels:     labels,
	}

	if err := s.cr.CreateBuild(ctx, capsuleID, b); err != nil {
		return "", err
	}

	return ref.Name(), nil
}

func (s *Service) ListBuilds(ctx context.Context, capsuleID uuid.UUID, pagination *model.Pagination) (iterator.Iterator[*capsule.Build], uint64, error) {
	return s.cr.ListBuilds(ctx, pagination, capsuleID)
}

func (s *Service) validateImage(ctx context.Context, ref name.Reference) (string, error) {
	if ok, d, err := s.cg.ImageExistsNatively(ctx, ref.String()); err != nil {
		return "", err
	} else if ok {
		return d, nil
	}

	var opts []remote.Option
	if ds, err := s.ps.GetProjectDockerSecret(ctx, ref.Context().RegistryStr()); errors.IsNotFound(err) {
	} else if err != nil {
		return "", err
	} else {
		opts = append(opts, remote.WithAuth(&authn.Basic{
			Username: ds.GetUsername(),
			Password: ds.GetPassword(),
		}))
	}

	img, err := remote.Image(ref, opts...)
	if err != nil {
		if terr, ok := err.(*transport.Error); ok {
			if len(terr.Errors) > 0 {
				switch terr.Errors[0].Code {
				case transport.UnauthorizedErrorCode:
					return "", errors.UnauthenticatedErrorf("error checking container registry '%s': %v", ref.Context().RegistryStr(), terr.Errors[0].Message)
				case transport.ManifestUnknownErrorCode:
					return "", errors.NotFoundErrorf("tag `%s` not found in container registry", ref.Identifier())
				default:
					return "", errors.UnknownErrorf("error from container registry '%s': %v", ref.Context().RegistryStr(), terr.Errors[0].String())
				}
			}
		}
		return "", err
	}

	d, err := img.Digest()
	if err != nil {
		return "", err
	}

	return d.String(), nil
}
