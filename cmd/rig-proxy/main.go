package main

import (
	"fmt"
	"os"
)

func main() {
	rootCmd := createRootCMD()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
