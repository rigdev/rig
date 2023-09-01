package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/volume"
	"go.uber.org/zap"
)

func (c *Client) CreateVolume(ctx context.Context, id string) error {
	logger := c.logger.With(zap.String("volume", id))
	logger.Info("creating volume")

	if _, err := c.dc.VolumeCreate(ctx, volume.CreateOptions{Name: fmt.Sprint(id)}); err != nil {
		return err
	}

	return nil
}
