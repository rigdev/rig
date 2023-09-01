package repository

import (
	"context"

	"github.com/rigdev/rig/pkg/uuid"
)

type Secret interface {
	Create(ctx context.Context, secretID uuid.UUID, secret []byte) error
	Update(ctx context.Context, secretID uuid.UUID, secret []byte) error
	Get(ctx context.Context, secretID uuid.UUID) ([]byte, error)
	Delete(ctx context.Context, secretID uuid.UUID) error
}
