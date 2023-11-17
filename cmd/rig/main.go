package main

import (
	"log"

	"github.com/rigdev/rig/cmd/rig/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
