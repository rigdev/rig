package s3

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rigdev/rig-go-api/api/v1/storage"
)

type Storage struct {
	s3 *s3.Client
}

func New(cfg *storage.S3Config) (*Storage, error) {
	credentials := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.GetCredentials().GetPublicKey(), cfg.GetCredentials().GetPrivateKey(), ""))
	config := aws.NewConfig()
	config.Credentials = credentials
	config.Region = cfg.Region

	client := s3.NewFromConfig(*config)
	return &Storage{
		s3: client,
	}, nil
}
