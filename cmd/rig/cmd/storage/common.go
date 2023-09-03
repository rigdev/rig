package storage

import (
	"net/url"

	"github.com/rigdev/rig/pkg/errors"
)

func isNSUri(raw string) bool {
	_, _, err := parseRSUri(raw)
	return err == nil
}

func parseRSUri(raw string) (string, string, error) {
	uri, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}

	if uri.Scheme != "rs" {
		return "", "", errors.InvalidArgumentErrorf("expect file of format `ns://bucket/path/to/file`")
	}

	return uri.Host, uri.Path, nil
}
