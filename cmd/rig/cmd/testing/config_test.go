package testinteractive

import (
	"testing"

	"github.com/rigdev/rig/pkg/cli"
)

func Test_stuff(t *testing.T) {
	cli.IsProduction = false
}
