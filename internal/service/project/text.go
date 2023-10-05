package project

import (
	"context"

	"github.com/rigdev/rig/pkg/gateway/text"
	"github.com/rigdev/rig/pkg/errors"
)

func (s *service) GetTextProvider(ctx context.Context) (text.Gateway, error) {
	return nil, errors.UnimplementedErrorf("GetTextProvider")
}
