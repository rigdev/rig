package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
)

func (s *Storage) DeleteObject(ctx context.Context, bucketName, path string) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}
	if err := s.minioClient.RemoveObject(ctx, bucketName, path, opts); err != nil {
		return err
	}
	return nil
}
