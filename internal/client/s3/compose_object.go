package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (s *Storage) ComposeObject(ctx context.Context, bucketName string, dest string, srcs ...string) error {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(dest),
	}

	output, err := s.s3.CreateMultipartUpload(ctx, input)
	if err != nil {
		return err
	}

	parts := make([]types.CompletedPart, len(srcs))
	for i, src := range srcs {
		input := &s3.UploadPartCopyInput{
			Bucket:     output.Bucket,
			Key:        output.Key,
			CopySource: aws.String(bucketName + src),
			PartNumber: int32(i),
			UploadId:   output.UploadId,
		}
		o, err := s.s3.UploadPartCopy(ctx, input)
		if err != nil {
			return err
		}
		part := types.CompletedPart{
			ETag:       o.CopyPartResult.ETag,
			PartNumber: int32(i),
		}
		parts = append(parts, part)
	}
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   output.Bucket,
		Key:      output.Key,
		UploadId: output.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: parts,
		},
	}
	_, err = s.s3.CompleteMultipartUpload(ctx, completeInput)
	if err != nil {
		return err
	}
	return nil
}
