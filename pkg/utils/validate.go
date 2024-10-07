package utils

import (
	"regexp"

	"github.com/rigdev/rig/pkg/errors"
)

func ValidateSystemName(input string) error {
	if l := len(input); l < 1 || l > 63 {
		return errors.InvalidArgumentErrorf("must be between 1 and 63 characters long")
	}

	if !regexp.MustCompile(`^[a-z][a-z0-9-_]+$`).MatchString(input) {
		return errors.InvalidArgumentErrorf("invalid name; can only contain a-z, 0-9, '-' and '_'")
	}

	return nil
}
