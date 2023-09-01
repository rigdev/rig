package gcs

import (
	"context"
	"encoding/json"
	"fmt"

	gStorage "cloud.google.com/go/storage"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"google.golang.org/api/option"
)

type Storage struct {
	gcsClient *gStorage.Client
	projectID string
}

func New(ctx context.Context, cfg *storage.GcsConfig) (*Storage, error) {
	c := map[string]interface{}{}
	if err := json.Unmarshal(cfg.GetConfig(), &c); err != nil {
		return nil, err
	}

	projectID, ok := c["project_id"].(string)
	if !ok {
		return nil, fmt.Errorf("project_id not found in credentials")
	}

	client, err := gStorage.NewClient(ctx, option.WithCredentialsJSON(cfg.GetConfig()))
	if err != nil {
		return nil, err
	}
	return &Storage{gcsClient: client, projectID: projectID}, nil
}
