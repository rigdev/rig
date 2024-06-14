package license

import (
	"math"

	"github.com/rigdev/rig-go-api/api/v1/settings"
)

type UserLimit int32

const (
	UnspecifiedNumUsers UserLimit = 2
	FreeNumUsers        UserLimit = 5
	TeamNumUsers        UserLimit = 10
	EnterpriseNumUsers  UserLimit = math.MaxInt32
)

func GetUserLimit(plan settings.Plan) UserLimit {
	switch plan {
	case settings.Plan_PLAN_FREE:
		return FreeNumUsers
	case settings.Plan_PLAN_TEAM:
		return TeamNumUsers
	case settings.Plan_PLAN_ENTERPRISE:
		return EnterpriseNumUsers
	default:
		return UnspecifiedNumUsers
	}
}
