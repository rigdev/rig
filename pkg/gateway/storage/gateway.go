package storage

import (
	"context"
	"io"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
)

type Gateway interface {
	Test(ctx context.Context) error

	UploadObject(ctx context.Context, reader io.Reader, size int64, bucket, path, contentType string) (string, uint64, error)
	DownloadObject(ctx context.Context, bucket, path string) (io.ReadSeekCloser, error)

	GetObject(ctx context.Context, bucket, path string) (*storage.Object, error)
	CopyObject(ctx context.Context, dstBucket, dstPath, srcBucket, srcPath string) error
	DeleteObject(ctx context.Context, bucket, path string) error
	ListObjects(ctx context.Context, bucketName, token, prefix,
		startpath, endpath string, recursive bool, limit uint32) (string, iterator.Iterator[*storage.ListObjectsResponse_Result], error)
	ComposeObject(ctx context.Context, bucketName string, dest string, srcs ...string) error

	CreateBucket(ctx context.Context, name, region string) (string, error)
	GetBucket(ctx context.Context, name string) (*storage.Bucket, error)
	DeleteBucket(ctx context.Context, name string) error
	ListBuckets(ctx context.Context) (iterator.Iterator[*storage.Bucket], error)
}
