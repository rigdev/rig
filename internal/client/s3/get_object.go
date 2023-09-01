package s3

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) GetObject(ctx context.Context, bucketName, path string) (*storage.Object, error) {
	input := s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(strings.TrimPrefix(path, "/")),
	}

	output, err := s.s3.GetObject(ctx, &input)
	if err != nil {
		return nil, err
	}

	return &storage.Object{
		Path:         path,
		LastModified: timestamppb.New(*output.LastModified),
		Size:         uint64(output.ContentLength),
		Etag:         *output.ETag,
		ContentType:  *output.ContentType,
	}, nil
}
