package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

func (s *Storage) CreateBucket(ctx context.Context, name, region string) (string, error) {
	exists, err := s.minioClient.BucketExists(ctx, name)
	if err != nil {
		return "", err
	}
	if exists {
		fmt.Println("bucket already exists, and is now linked to project")
		return name, nil
	} else {
		return name, toError(s.minioClient.MakeBucket(ctx, name, minio.MakeBucketOptions{Region: region, ObjectLocking: false}))
	}
}
