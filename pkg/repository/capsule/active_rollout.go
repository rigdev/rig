package capsule

import (
	"time"
)

type ActiveRollout struct {
	ProjectID   string
	CapsuleID   string
	RolloutID   uint64
	ScheduledAt time.Time
}
