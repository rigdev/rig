package storage

import (
	"net/url"

	"github.com/rigdev/rig/pkg/errors"
)

func isRigUri(raw string) bool {
	_, _, err := parseRigUri(raw)
	return err == nil
}

func parseRigUri(raw string) (string, string, error) {
	uri, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}

	if uri.Scheme != "rig" {
		return "", "", errors.InvalidArgumentErrorf("expect file of format `rig://bucket/path/to/file`")
	}

	return uri.Host, uri.Path, nil
}
