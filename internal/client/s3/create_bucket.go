package s3

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *Storage) CreateBucket(ctx context.Context, name, region string) (string, error) {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}

	if region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	_, err := s.s3.CreateBucket(ctx, input)
	if err != nil {
		var alreadyExists *types.BucketAlreadyOwnedByYou
		if errors.As(err, &alreadyExists) {
			return name, nil
		}
		return "", err
	}

	return name, nil
}
