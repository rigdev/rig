package segment

import (
	"github.com/rigdev/rig/internal/config"
	"github.com/segmentio/analytics-go/v3"
	"go.uber.org/zap"
)

type Client struct {
	client  analytics.Client
	appInfo analytics.AppInfo
	logger  *zap.Logger
}

const _writeKey = "v00FCszKOcE7YFHOLUHCgoGdam9QhvRi"

// New implements text.Provider interface using the Twilio client.
func New(cfg config.Config, logger *zap.Logger) *Client {
	if cfg.Telemetry.Enabled {
		logger.Info("admin usage telemetry enabled. see https://docs.rig.dev/usage for more information.")
	}

	return &Client{
		client: analytics.New(_writeKey),
		appInfo: analytics.AppInfo{
			Name: "rig",
		},
		logger: logger,
	}
}

func (c *Client) Track(msg analytics.Track) {
	msg.Context = c.context()
	if err := c.client.Enqueue(msg); err != nil {
		c.logger.Debug("invalid segment message", zap.Error(err))
	}
}

func (c *Client) Identify(msg analytics.Identify) {
	msg.Context = c.context()
	if err := c.client.Enqueue(msg); err != nil {
		c.logger.Debug("invalid segment message", zap.Error(err))
	}
}

func (c *Client) Alias(msg analytics.Alias) {
	msg.Context = c.context()
	if err := c.client.Enqueue(msg); err != nil {
		c.logger.Debug("invalid segment message", zap.Error(err))
	}
}

func (c *Client) context() *analytics.Context {
	return &analytics.Context{
		App: c.appInfo,
	}
}
