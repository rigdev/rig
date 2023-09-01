package database

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/iterator"
)

func (s *Service) List(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*database.Database], uint64, error) {
	it, total, err := s.dr.List(ctx, pagination)
	if err != nil {
		return nil, 0, err
	}
	return it, total, nil
}
