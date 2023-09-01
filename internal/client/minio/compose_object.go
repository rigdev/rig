package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
)

func (s *Storage) ComposeObject(ctx context.Context, bucketName string, dest string, srcs ...string) error {
	var ss []minio.CopySrcOptions
	for _, p := range srcs {
		ss = append(ss, minio.CopySrcOptions{
			Bucket: bucketName,
			Object: toPath(p),
		})
	}

	if _, err := s.minioClient.ComposeObject(ctx, minio.CopyDestOptions{
		Bucket: bucketName,
		Object: toPath(dest),
	}, ss...); err != nil {
		return err
	}

	return nil
}
