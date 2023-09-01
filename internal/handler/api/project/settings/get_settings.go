package settings

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project/settings"
)

func (h *Handler) GetSettings(ctx context.Context, req *connect.Request[settings.GetSettingsRequest]) (*connect.Response[settings.GetSettingsResponse], error) {
	res, err := h.ps.GetProjectSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &connect.Response[settings.GetSettingsResponse]{
		Msg: &settings.GetSettingsResponse{
			Settings: res,
		},
	}, nil
}
