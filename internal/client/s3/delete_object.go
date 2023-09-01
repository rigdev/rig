package s3

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *Storage) DeleteObject(ctx context.Context, bucket, path string) error {
	input := s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(strings.TrimPrefix(path, "/")),
	}
	_, err := s.s3.DeleteObject(ctx, &input)
	if err != nil {
		return err
	}

	return nil
}
