package base

import (
	"strings"

	"github.com/spf13/cobra"
)

type PromptInformation struct {
	ContextCreation bool
}

const (
	RFC3339NanoFixed  = "2006-01-02T15:04:05.000000000Z07:00"
	RFC3339MilliFixed = "2006-01-02T15:04:05.000Z07:00"
)

func cmdPathContainsUsePrefix(cmd *cobra.Command, use string) bool {
	for cmd := cmd; cmd != nil; cmd = cmd.Parent() {
		if strings.HasPrefix(cmd.Use, use) {
			return true
		}
	}
	return false
}

func SkipChecks(cmd *cobra.Command) bool {
	return cmdPathContainsUsePrefix(cmd, "completion") || cmdPathContainsUsePrefix(cmd, "help ")
}
