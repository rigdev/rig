package gcs

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) GetBucket(ctx context.Context, name string) (*storage.Bucket, error) {
	bkt := s.gcsClient.Bucket(name)
	attrs, err := bkt.Attrs(ctx)
	if err != nil {
		return nil, err
	}

	return &storage.Bucket{
		Region:         attrs.Location,
		ProviderBucket: attrs.Name,
		CreatedAt:      timestamppb.New(attrs.Created),
	}, nil
}
