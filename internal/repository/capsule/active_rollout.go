package capsule

import (
	"time"

	"github.com/rigdev/rig/pkg/uuid"
)

type ActiveRollout struct {
	ProjectID   uuid.UUID
	CapsuleID   string
	RolloutID   uint64
	ScheduledAt time.Time
}
