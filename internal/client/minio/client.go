// minio implements the storage.Provider interface using the Minio client.
package minio

import (
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/errors"
)

type Storage struct {
	minioClient *minio.Client
}

func NewDefault(cfg config.Config) (*Storage, error) {
	minioCfg := &storage.MinioConfig{
		Credentials: &model.ProviderCredentials{
			PublicKey:  cfg.Client.Minio.AccessKeyID,
			PrivateKey: cfg.Client.Minio.SecretAccessKey,
		},
		Endpoint: cfg.Client.Minio.Host,
		Secure:   cfg.Client.Minio.Secure,
	}

	return New(minioCfg)
}

func New(cfg *storage.MinioConfig) (*Storage, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV2(cfg.Credentials.GetPublicKey(), cfg.Credentials.GetPrivateKey(), ""),
		Secure: cfg.Secure,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return &Storage{
		minioClient: minioClient,
	}, nil
}

func toError(err error) error {
	if err == nil {
		return nil
	}

	er := minio.ToErrorResponse(err)
	if er.Key != "" && er.StatusCode == http.StatusNotFound {
		return errors.NotFoundErrorf("object not found")
	}
	return errors.FromHTTP(er.StatusCode, er.Message)
}

func toPath(path string) string {
	return strings.TrimPrefix(path, "/")
}
