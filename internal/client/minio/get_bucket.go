package minio

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) GetBucket(ctx context.Context, name string) (*storage.Bucket, error) {
	minioBuckets, err := s.minioClient.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	var bucket *storage.Bucket
	for _, minioBucket := range minioBuckets {
		if minioBucket.Name == name {
			bucket = &storage.Bucket{
				Name:      name,
				CreatedAt: timestamppb.New(minioBucket.CreationDate),
			}
			break
		}
	}

	if bucket == nil {
		return nil, errors.NotFoundErrorf("bucket %s not found", name)
	}

	region, err := s.minioClient.GetBucketLocation(ctx, name)
	if err != nil {
		return nil, err
	}

	bucket.Region = region
	return bucket, nil
}
