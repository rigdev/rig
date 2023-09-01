package s3

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *Storage) CopyObject(ctx context.Context, dstBucket, dstPath, srcBucket, srcPath string) error {
	input := s3.CopyObjectInput{
		Bucket: aws.String(dstBucket),
		Key:    aws.String(strings.TrimPrefix(dstPath, "/")),
		CopySource: aws.String(
			srcBucket + srcPath,
		),
	}

	_, err := s.s3.CopyObject(ctx, &input)
	if err != nil {
		return err
	}

	return nil
}
