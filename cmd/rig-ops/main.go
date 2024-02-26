package main

import (
	"log"

	"github.com/rigdev/rig/cmd/rig-ops/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
