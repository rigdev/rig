package main

import (
	"fmt"
	"os"

	"github.com/rigdev/rig/cmd/rig-ops/cmd"
	"github.com/rigdev/rig/pkg/errors"
	"go.uber.org/dig"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Println(errors.MessageOf(dig.RootCause(err)))
		os.Exit(1)
	}
}
