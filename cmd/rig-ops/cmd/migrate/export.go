package migrate

import (
	"os"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/cmd/common"
)

func exportCapsule(capsule *platformv1.Capsule, path string) error {
	str, err := common.Format(capsule, common.OutputTypeYAML)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(str), 0644)
}
