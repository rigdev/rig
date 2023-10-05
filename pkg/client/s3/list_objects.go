package s3

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) ListObjects(ctx context.Context, bucketName, token, prefix,
	startpath, endpath string, recursive bool, limit uint32,
) (string, iterator.Iterator[*storage.ListObjectsResponse_Result], error) {
	input := &s3.ListObjectsV2Input{
		Bucket:     aws.String(bucketName),
		MaxKeys:    int32(limit),
		Prefix:     aws.String(strings.TrimPrefix(prefix, "/")),
		StartAfter: aws.String(startpath),
	}
	if token != "" {
		input.ContinuationToken = aws.String(token)
	}

	if !recursive {
		input.Delimiter = aws.String("/")
	}

	res, err := s.s3.ListObjectsV2(ctx, input)
	if err != nil {
		return "", nil, err
	}

	oit := iterator.NewProducer[*storage.ListObjectsResponse_Result]()
	continuationToken := ""
	if res.IsTruncated {
		continuationToken = *res.NextContinuationToken
	}

	go func() {
		defer oit.Done()
		for _, dir := range res.CommonPrefixes {
			rigObj := &storage.ListObjectsResponse_Result{
				Result: &storage.ListObjectsResponse_Result_Folder{
					Folder: strings.TrimPrefix(*dir.Prefix, "/"),
				},
			}
			err := oit.Value(rigObj)
			if err != nil {
				oit.Error(err)
				return
			}
		}
		for _, obj := range res.Contents {
			rigObj, err := objToNunt(obj, endpath, prefix)
			if err != nil {
				oit.Error(err)
				return
			}
			if rigObj != nil {
				err = oit.Value(rigObj)
				if err != nil {
					oit.Error(err)
					return
				}
			}
		}
	}()

	return continuationToken, oit, nil
}

func objToNunt(obj types.Object, endpath, prefix string) (*storage.ListObjectsResponse_Result, error) {
	if endpath != "" && strings.Compare(*obj.Key, endpath) >= 0 {
		return nil, io.EOF
	}

	key := "/" + *obj.Key
	if strings.HasSuffix(key, "/") {
		return nil, nil
	}

	return &storage.ListObjectsResponse_Result{
		Result: &storage.ListObjectsResponse_Result_Object{
			Object: &storage.Object{
				Path:         key,
				LastModified: timestamppb.New(*obj.LastModified),
				Size:         uint64(obj.Size),
				Etag:         *obj.ETag,
			},
		},
	}, nil
}
