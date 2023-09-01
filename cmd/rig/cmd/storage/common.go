package storage

import (
	"net/url"

	"github.com/rigdev/rig/pkg/errors"
)

func isNSUri(raw string) bool {
	_, _, err := parseNSUri(raw)
	return err == nil
}

func parseNSUri(raw string) (string, string, error) {
	uri, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}

	if uri.Scheme != "ns" {
		return "", "", errors.InvalidArgumentErrorf("expect file of format `ns://bucket/path/to/file`")
	}

	return uri.Host, uri.Path, nil
}
