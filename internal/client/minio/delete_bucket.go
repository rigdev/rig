package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
)

func (s *Storage) DeleteBucket(ctx context.Context, name string) error {
	exist, err := s.minioClient.BucketExists(ctx, name)
	if err != nil {
		return err
	}
	if exist {
		if err := s.minioClient.RemoveBucketWithOptions(ctx, name, minio.RemoveBucketOptions{ForceDelete: true}); err != nil {
			return err
		}
	}
	return nil
}
