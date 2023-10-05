package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
)

func (m *Storage) CopyObject(ctx context.Context, dstBucket, dstPath, srcBucket, srcPath string) error {

	dstOpts := minio.CopyDestOptions{
		Bucket: dstBucket,
		Object: dstPath,
	}

	srcOpts := minio.CopySrcOptions{
		Bucket: srcBucket,
		Object: srcPath,
	}

	if _, err := m.minioClient.CopyObject(ctx, dstOpts, srcOpts); err != nil {
		return err
	}
	return nil
}
