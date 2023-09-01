package minio

import "context"

// Ping checks if the minio client is able to connect.
func (s *Storage) Test(ctx context.Context) error {
	_, err := s.minioClient.ListBuckets(ctx)
	if err != nil {
		return err
	}
	return nil
}
