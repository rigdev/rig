package gcs

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
)

func (s *Storage) CreateBucket(ctx context.Context, name, region string) (string, error) {
	bkt := s.gcsClient.Bucket(name)
	_, err := bkt.Attrs(ctx)
	if err != nil {
		if err := bkt.Create(ctx, s.projectID, &storage.BucketAttrs{
			Location: region,
		}); err != nil {
			return "", err
		}
	} else {
		fmt.Println("Bucket already exists, and is now linked")
		return name, nil
	}

	return name, nil
}
