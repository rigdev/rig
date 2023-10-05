package gcs

import (
	"context"

	"google.golang.org/api/iterator"
)

// Test checks if the GCS client is able to connect to the GCS API.
func (s *Storage) Test(ctx context.Context) error {
	bit := s.gcsClient.Buckets(ctx, s.projectID)
	_, err := bit.Next()
	if err != nil && err != iterator.Done {
		return err
	}

	return nil
}
