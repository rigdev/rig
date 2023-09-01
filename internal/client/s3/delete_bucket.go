package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *Storage) DeleteBucket(ctx context.Context, name string) error {
	input := s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}

	_, err := s.s3.DeleteBucket(ctx, &input)
	if err != nil {
		return err
	}

	return nil
}
