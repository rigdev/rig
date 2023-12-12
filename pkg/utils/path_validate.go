package utils

import (
	"fmt"
	"regexp"
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

// ValidateURLPath validates the Path segment of an URL is correct
// https://datatracker.ietf.org/doc/html/rfc3986#section-3.3
func ValidateURLPath(p string) error {
	if len(p) == 0 {
		return nil
	}

	if len(p) >= 2 && p[:2] == "//" {
		return errors.New("path cannot start with '//'")
	}

	hexEscapeRegex := "%[0-9a-fA-F]{2}"
	pcharRegex := fmt.Sprintf("[A-Za-z0-9:@\\-._~]|(%s)", hexEscapeRegex)
	urlRegex := fmt.Sprintf("(/(%s)*)+", pcharRegex)

	r, err := regexp.Compile(urlRegex)
	if err != nil {
		return err
	}
	r.Longest()
	match := r.FindString(p)
	if len(p) != len(match) {
		return errors.New("url path is malformed")
	}

	return nil
}
