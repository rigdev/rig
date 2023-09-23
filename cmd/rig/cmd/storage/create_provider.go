package storage

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) createProvider(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	var err error
	if name == "" {
		name, err = common.PromptInput("Provider identifier:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	var config *storage.Config
	var providerType string
	if GCS {
		providerType = "Google Cloud Storage"
		config, err = getGCSConfig(credsFilePath)
		if err != nil {
			return err
		}
	} else if S3 {
		providerType = "Amazon S3"
		config = &storage.Config{
			Config: &storage.Config_S3{
				S3: &storage.S3Config{
					Credentials: &model.ProviderCredentials{
						PublicKey:  accessKey,
						PrivateKey: secretKey,
					},
					Region: region,
				},
			},
		}
	} else if Minio {
		providerType = "Minio"
		config = &storage.Config{
			Config: &storage.Config_Minio{
				Minio: &storage.MinioConfig{
					Endpoint: endpoint,
					Credentials: &model.ProviderCredentials{
						PublicKey:  accessKey,
						PrivateKey: secretKey,
					},
				},
			},
		}
	} else {
		fields := []string{
			"Google Cloud Storage",
			"Amazon S3",
			"Minio",
		}
		var i int
		i, providerType, err = common.PromptSelect("Provider type:", fields)
		if err != nil {
			return err
		}

		switch i {
		case 0:
			// GCS
			path, err := common.PromptInput("Credentials Path:", common.ValidateNonEmptyOpt)
			if err != nil {
				return err
			}

			config, err = getGCSConfig(path)
			if err != nil {
				return err
			}
		case 1:
			// S3
			accessKey, err := common.PromptInput("Access Key:", common.ValidateNonEmptyOpt)
			if err != nil {
				return err
			}

			secretKey, err := common.PromptInput("Secret Key:", common.ValidateNonEmptyOpt)
			if err != nil {
				return err
			}

			region, err := common.PromptInput("Region:", common.ValidateNonEmptyOpt)
			if err != nil {
				return err
			}

			config = &storage.Config{
				Config: &storage.Config_S3{
					S3: &storage.S3Config{
						Credentials: &model.ProviderCredentials{
							PublicKey:  accessKey,
							PrivateKey: secretKey,
						},
						Region: region,
					},
				},
			}

		case 2:
			// Minio

			accessKey, err := common.PromptInput("Access Key:", common.ValidateNonEmptyOpt)
			if err != nil {
				return err
			}

			secretKey, err := common.PromptInput("Secret Key:", common.ValidateNonEmptyOpt)
			if err != nil {
				return err
			}

			endpoint, err := common.PromptInput("Endpoint:", common.ValidateNonEmptyOpt)
			if err != nil {
				return err
			}

			config = &storage.Config{
				Config: &storage.Config_Minio{
					Minio: &storage.MinioConfig{
						Endpoint: endpoint,
						Credentials: &model.ProviderCredentials{
							PublicKey:  accessKey,
							PrivateKey: secretKey,
						},
					},
				},
			}
		}
	}

	_, err = c.Rig.Storage().CreateProvider(ctx, &connect.Request[storage.CreateProviderRequest]{
		Msg: &storage.CreateProviderRequest{
			Name:        name,
			Config:      config,
			LinkBuckets: linkBuckets,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("%s provder %s created", providerType, name))
	return nil
}

func getGCSConfig(path string) (*storage.Config, error) {
	// load json credentials file from path
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, stat.Size())
	_, err = bufio.NewReader(f).Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &storage.Config{
		Config: &storage.Config_Gcs{
			Gcs: &storage.GcsConfig{
				Config: buf,
			},
		},
	}, nil
}
