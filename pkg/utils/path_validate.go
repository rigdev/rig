package utils

import (
	"strings"

	"github.com/rigdev/rig/pkg/errors"
)

// valid path:
// 1. absolute
// 2. no trailing slash
// 3. no not escaped whitespace
// 4. no double slashes
// 5. no dots
// 6. non-empty
func ValiateConfigFilePath(p string) error {
	if p == "" {
		return errors.InvalidArgumentErrorf("must not be empty")
	}

	segments := strings.Split(p, "/")
	if segments[0] != "" {
		return errors.InvalidArgumentErrorf("must be an absolute path")
	}
	for i := 1; i < len(segments); i++ {
		s := segments[i]
		// check for unescaped whitespace
		if j := strings.Index(s, " "); j != -1 && s[j-1] != '\\' {
			return errors.InvalidArgumentErrorf("must not contain unescaped whitespace")
		}
		if s == "" {
			if i == len(segments)-1 {
				return errors.InvalidArgumentErrorf("must not end with a slash")
			}
			return errors.InvalidArgumentErrorf("must not contain double slashes")
		}
		if s == "." || s == ".." {
			return errors.InvalidArgumentErrorf("must not contain dots")
		}
	}

	return nil
}
