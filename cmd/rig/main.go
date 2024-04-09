package main

import (
	"fmt"
	"os"

	"github.com/rigdev/rig/cmd/rig/cmd"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/errors"
	"go.uber.org/dig"
)

func main() {
	if err := cmd.Run(cli.StandardSetupContext); err != nil {
		fmt.Println(errors.MessageOf(dig.RootCause(err)))
		os.Exit(1)
	}
}
