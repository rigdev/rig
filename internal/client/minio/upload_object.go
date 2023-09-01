package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

func (s *Storage) UploadObject(ctx context.Context, reader io.Reader, size int64, bucketName, path, contentType string) (string, uint64, error) {
	info, err := s.minioClient.PutObject(ctx, bucketName, path, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", 0, err
	}

	return path, uint64(info.Size), nil
}
