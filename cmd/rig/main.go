package main

import (
	"fmt"
	"os"

	"github.com/rigdev/rig/cmd/rig/cmd"
	"github.com/rigdev/rig/pkg/errors"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Println(errors.MessageOf(err))
		os.Exit(1)
	}
}
