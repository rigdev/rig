package capsule

import (
	"time"

	"github.com/rigdev/rig/pkg/uuid"
)

type ActiveRollout struct {
	ProjectID   uuid.UUID
	CapsuleID   uuid.UUID
	RolloutID   uint64
	ScheduledAt time.Time
}
