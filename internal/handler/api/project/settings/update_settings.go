package settings

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project/settings"
)

func (h *Handler) UpdateSettings(ctx context.Context, req *connect.Request[settings.UpdateSettingsRequest]) (*connect.Response[settings.UpdateSettingsResponse], error) {
	err := h.ps.UpdateProjectSettings(ctx, req.Msg.GetUpdates())
	if err != nil {
		return nil, err
	}
	return &connect.Response[settings.UpdateSettingsResponse]{
		Msg: &settings.UpdateSettingsResponse{},
	}, nil
}
